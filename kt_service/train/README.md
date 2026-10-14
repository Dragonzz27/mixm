# DKT 公开数据训练说明

## 1）准备公开数据 CSV

输入 CSV 需要包含以下字段：

- `user_id`
- `skill_name`
- `correct`（0/1）
- `timestamp`（整数时间序）

示例路径：

`kt_service/train/data/public_raw.csv`

## 2）预处理：生成序列与知识点映射

```powershell
py kt_service/train/prepare_public_dataset.py `
  --input-csv kt_service/train/data/public_raw.csv `
  --output-dir kt_service/train/output
```

输出文件：

- `kt_service/train/output/sequences.json`
- `kt_service/train/output/skill_map.json`

## 3）训练 DKT 并导出 checkpoint

```powershell
py kt_service/train/train_dkt_checkpoint.py `
  --prepared-dir kt_service/train/output `
  --out-checkpoint kt_service/train/output/dkt_checkpoint.pt `
  --epochs 5 `
  --batch-size 32 `
  --max-len 100
```

## 4）让 kt-service 加载训练结果

需要提供两个文件：

- `dkt_checkpoint.pt`
- `skill_map.json`

在项目根目录的 `deploy/docker-compose.yaml` 中为 `kt-service` 配置（以下 Compose 命令从项目根目录执行）：

- 模型路径：`KT_DKT_CHECKPOINT=.../dkt_checkpoint.pt`
- 映射路径：`KT_DKT_SKILL_MAP=.../skill_map.json`
- 并把宿主机目录挂载到容器内（`volumes`）

完成后重启：

```powershell
docker compose -f deploy/docker-compose.yaml up -d --build kt-service
```

## 5）快速验证

1. 查看健康检查：`http://localhost:8010/health`
2. 提交一次模拟测试
3. 查看日志：

```powershell
docker compose -f deploy/docker-compose.yaml logs kt-service --tail 100
```

如果看到 `POST /predict 200`，说明 KT 推理链路正常。
