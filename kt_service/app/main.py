from __future__ import annotations

import json
import os
from collections import defaultdict
from dataclasses import dataclass
from typing import Dict, List, Tuple

import torch
import torch.nn as nn
from fastapi import FastAPI
from pydantic import BaseModel, Field


class TraceItem(BaseModel):
    question_id: str
    knowledge_tag: str
    is_correct: bool


class PredictRequest(BaseModel):
    user_id: str
    traces: List[TraceItem] = Field(default_factory=list)
    prev_mastery: Dict[str, float] = Field(default_factory=dict)


class PredictResponse(BaseModel):
    model: str
    mastery_by_tag: Dict[str, float]


app = FastAPI(title="KT Service", version="1.0.0")


@dataclass
class SeqSample:
    skill_ids: List[int]
    responses: List[int]
    tag_order: List[str]


class DKTNet(nn.Module):
    """A compact DKT-style network for sequence inference."""

    def __init__(self, max_skills: int = 4096, emb_dim: int = 32, hidden_dim: int = 64):
        super().__init__()
        self.max_skills = max_skills
        self.skill_emb = nn.Embedding(max_skills + 1, emb_dim, padding_idx=0)
        self.inter_emb = nn.Embedding((max_skills * 2) + 2, emb_dim, padding_idx=0)
        self.gru = nn.GRU(input_size=emb_dim, hidden_size=hidden_dim, batch_first=True)
        self.head = nn.Linear(hidden_dim, max_skills + 1)

    def forward(self, inter_seq: torch.Tensor) -> torch.Tensor:
        x = self.inter_emb(inter_seq)
        h, _ = self.gru(x)
        logits = self.head(h)
        return torch.sigmoid(logits)


class DKTService:
    """
    DKT runtime adapter:
    - maps knowledge tags to integer IDs
    - builds sequence tensors + padding mask
    - runs torch forward pass
    """

    def __init__(self):
        self.tag2id: Dict[str, int] = {}
        self.id2tag: Dict[int, str] = {}
        self.max_skills = 4096
        self.max_seq_len = 200
        self.model = DKTNet(max_skills=self.max_skills)
        self._load_checkpoint_if_exists()
        self.model.eval()

    def _load_checkpoint_if_exists(self) -> None:
        checkpoint_path = os.getenv("KT_DKT_CHECKPOINT", "").strip()
        skill_map_path = os.getenv("KT_DKT_SKILL_MAP", "").strip()
        if not checkpoint_path:
            return
        if not os.path.exists(checkpoint_path):
            return

        if skill_map_path and os.path.exists(skill_map_path):
            with open(skill_map_path, "r", encoding="utf-8") as f:
                loaded = json.load(f)
            # file stores: {"skill_name": 1, ...}
            self.tag2id = {str(k): int(v) for k, v in loaded.items()}
            self.id2tag = {int(v): str(k) for k, v in self.tag2id.items()}

        checkpoint = torch.load(checkpoint_path, map_location="cpu")
        ckpt_max_skills = int(checkpoint.get("max_skills", self.max_skills))
        if ckpt_max_skills != self.max_skills:
            self.max_skills = ckpt_max_skills
            self.model = DKTNet(max_skills=self.max_skills)
        state_dict = checkpoint.get("state_dict", checkpoint)
        self.model.load_state_dict(state_dict, strict=False)

    def _ensure_skill_id(self, tag: str) -> int:
        if tag in self.tag2id:
            return self.tag2id[tag]
        next_id = len(self.tag2id) + 1
        if next_id > self.max_skills:
            # Clamp overflow tags into last bucket.
            return self.max_skills
        self.tag2id[tag] = next_id
        self.id2tag[next_id] = tag
        return next_id

    def _serialize(self, traces: List[TraceItem]) -> SeqSample:
        skill_ids: List[int] = []
        responses: List[int] = []
        tag_order: List[str] = []
        for t in traces[-self.max_seq_len :]:
            sid = self._ensure_skill_id(t.knowledge_tag)
            skill_ids.append(sid)
            responses.append(1 if t.is_correct else 0)
            tag_order.append(t.knowledge_tag)
        return SeqSample(skill_ids=skill_ids, responses=responses, tag_order=tag_order)

    def _build_batch_tensors(self, sample: SeqSample) -> Tuple[torch.Tensor, torch.Tensor, torch.Tensor]:
        seq_len = len(sample.skill_ids)
        if seq_len == 0:
            # Keep non-empty tensors to simplify downstream.
            return (
                torch.zeros((1, 1), dtype=torch.long),
                torch.zeros((1, 1), dtype=torch.long),
                torch.zeros((1, 1), dtype=torch.bool),
            )

        skills = torch.tensor([sample.skill_ids], dtype=torch.long)  # [B=1, T]
        rs = torch.tensor([sample.responses], dtype=torch.long)  # [B=1, T]
        # DKT interaction index: q + r * num_skills
        inter = skills + (rs * self.max_skills)
        mask = torch.ones((1, seq_len), dtype=torch.bool)
        return skills, inter, mask

    def _predict_tag_mastery(self, sample: SeqSample, prev_mastery: Dict[str, float]) -> Dict[str, float]:
        skills, inter, mask = self._build_batch_tensors(sample)
        with torch.no_grad():
            probs = self.model(inter)  # [1, T, S+1]

        # Aggregate predictions by knowledge tag from valid sequence positions.
        per_tag_values: Dict[str, List[float]] = defaultdict(list)
        if mask.any():
            valid_indices = mask[0].nonzero(as_tuple=False).flatten().tolist()
            for idx in valid_indices:
                sid = int(skills[0, idx].item())
                tag = self.id2tag.get(sid)
                if not tag:
                    continue
                p_model = float(probs[0, idx, sid].item())
                # Blend with previous mastery for stability.
                p_prev = float(prev_mastery.get(tag, 0.5))
                p = (0.65 * p_model) + (0.35 * p_prev)
                per_tag_values[tag].append(max(0.0, min(1.0, p)))

        result: Dict[str, float] = {}
        for tag, vals in per_tag_values.items():
            result[tag] = round(sum(vals) / max(len(vals), 1), 4)
        # Keep untouched tags from previous mastery.
        for tag, val in prev_mastery.items():
            if tag not in result:
                result[tag] = round(float(val), 4)
        return result

    def predict(self, traces: List[TraceItem], prev_mastery: Dict[str, float]) -> Dict[str, float]:
        sample = self._serialize(traces)
        if not sample.skill_ids:
            return {k: round(float(v), 4) for k, v in (prev_mastery or {}).items()}
        return self._predict_tag_mastery(sample, prev_mastery or {})


dkt_service = DKTService()


@app.get("/health")
def health():
    return {"ok": True, "model": "dkt", "mapped_skills": len(dkt_service.tag2id)}


@app.post("/predict", response_model=PredictResponse)
def predict(payload: PredictRequest):
    mastery = dkt_service.predict(payload.traces, payload.prev_mastery)
    return PredictResponse(model="dkt", mastery_by_tag=mastery)
