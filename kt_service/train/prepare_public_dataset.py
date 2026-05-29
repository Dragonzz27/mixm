import argparse
import csv
import json
from pathlib import Path


def build_skill_map(rows):
    skills = sorted({row["skill_name"] for row in rows if row.get("skill_name")})
    return {name: idx + 1 for idx, name in enumerate(skills)}


def group_by_user(rows):
    grouped = {}
    for row in rows:
        uid = row["user_id"]
        grouped.setdefault(uid, []).append(row)
    for uid in grouped:
        grouped[uid].sort(key=lambda x: int(x.get("timestamp", 0)))
    return grouped


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input-csv", required=True, help="raw public csv path")
    parser.add_argument("--output-dir", required=True, help="prepared output folder")
    args = parser.parse_args()

    input_csv = Path(args.input_csv)
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    rows = []
    with input_csv.open("r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for r in reader:
            # Required columns: user_id, skill_name, correct, timestamp
            user_id = r.get("user_id") or r.get("\ufeffuser_id")
            skill_name = r.get("skill_name") or r.get("\ufeffskill_name")
            if not user_id or not skill_name:
                continue
            rows.append(
                {
                    "user_id": str(user_id),
                    "skill_name": str(skill_name),
                    "correct": int(float(r.get("correct", 0))),
                    "timestamp": int(float(r.get("timestamp", 0))),
                }
            )

    skill_map = build_skill_map(rows)
    grouped = group_by_user(rows)

    sequences = []
    for _, items in grouped.items():
        skills = [skill_map[item["skill_name"]] for item in items]
        responses = [1 if item["correct"] > 0 else 0 for item in items]
        if len(skills) >= 2:
            sequences.append({"skills": skills, "responses": responses})

    (output_dir / "sequences.json").write_text(
        json.dumps(sequences, ensure_ascii=False), encoding="utf-8"
    )
    (output_dir / "skill_map.json").write_text(
        json.dumps(skill_map, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    print(f"prepared users={len(sequences)} skills={len(skill_map)}")


if __name__ == "__main__":
    main()
