import argparse
import json
from pathlib import Path
from typing import List

import torch
import torch.nn as nn
from torch.utils.data import DataLoader, Dataset


class DKTNet(nn.Module):
    def __init__(self, max_skills: int, emb_dim: int = 32, hidden_dim: int = 64):
        super().__init__()
        self.max_skills = max_skills
        self.inter_emb = nn.Embedding((max_skills * 2) + 2, emb_dim, padding_idx=0)
        self.gru = nn.GRU(input_size=emb_dim, hidden_size=hidden_dim, batch_first=True)
        self.head = nn.Linear(hidden_dim, max_skills + 1)

    def forward(self, inter_seq: torch.Tensor) -> torch.Tensor:
        x = self.inter_emb(inter_seq)
        h, _ = self.gru(x)
        return torch.sigmoid(self.head(h))


class SeqDataset(Dataset):
    def __init__(self, sequences: List[dict], max_len: int, max_skills: int):
        self.max_len = max_len
        self.max_skills = max_skills
        self.data = []
        for s in sequences:
            skills = s["skills"][-max_len:]
            rs = s["responses"][-max_len:]
            if len(skills) < 2:
                continue
            self.data.append((skills, rs))

    def __len__(self):
        return len(self.data)

    def __getitem__(self, idx):
        return self.data[idx]


def collate(batch, max_skills: int):
    max_t = max(len(x[0]) for x in batch)
    inter = torch.zeros((len(batch), max_t), dtype=torch.long)
    skill = torch.zeros((len(batch), max_t), dtype=torch.long)
    resp = torch.zeros((len(batch), max_t), dtype=torch.float32)
    mask = torch.zeros((len(batch), max_t), dtype=torch.bool)

    for i, (skills, rs) in enumerate(batch):
        t = len(skills)
        s = torch.tensor(skills, dtype=torch.long)
        r = torch.tensor(rs, dtype=torch.long)
        inter[i, :t] = s + (r * max_skills)
        skill[i, :t] = s
        resp[i, :t] = r.float()
        mask[i, :t] = True
    return inter, skill, resp, mask


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--prepared-dir", required=True)
    parser.add_argument("--out-checkpoint", required=True)
    parser.add_argument("--epochs", type=int, default=5)
    parser.add_argument("--batch-size", type=int, default=32)
    parser.add_argument("--max-len", type=int, default=100)
    args = parser.parse_args()

    prepared_dir = Path(args.prepared_dir)
    sequences = json.loads((prepared_dir / "sequences.json").read_text(encoding="utf-8"))
    skill_map = json.loads((prepared_dir / "skill_map.json").read_text(encoding="utf-8"))
    max_skills = max(int(v) for v in skill_map.values()) if skill_map else 1

    ds = SeqDataset(sequences, max_len=args.max_len, max_skills=max_skills)
    dl = DataLoader(
        ds,
        batch_size=args.batch_size,
        shuffle=True,
        collate_fn=lambda b: collate(b, max_skills=max_skills),
    )

    model = DKTNet(max_skills=max_skills)
    optim = torch.optim.Adam(model.parameters(), lr=1e-3)
    bce = nn.BCELoss(reduction="none")

    for epoch in range(args.epochs):
        model.train()
        total_loss = 0.0
        for inter, skill, resp, mask in dl:
            pred = model(inter)
            # predict current skill correctness at each step
            gather = pred.gather(dim=2, index=skill.unsqueeze(-1)).squeeze(-1)
            loss_raw = bce(gather, resp)
            loss = (loss_raw * mask.float()).sum() / mask.float().sum().clamp_min(1.0)
            optim.zero_grad()
            loss.backward()
            optim.step()
            total_loss += float(loss.item())
        avg = total_loss / max(len(dl), 1)
        print(f"epoch={epoch + 1} loss={avg:.4f}")

    out_path = Path(args.out_checkpoint)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    torch.save({"state_dict": model.state_dict(), "max_skills": max_skills}, out_path)
    print(f"saved checkpoint: {out_path}")


if __name__ == "__main__":
    main()

