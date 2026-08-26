# 实验室受控材料退役销毁见证台

这是一个面向实验室安全管理员、材料责任人、现场见证人员和合规复核人员的浏览器工作台。系统把受控材料从退役建档推进到双人清点、风险审查、销毁见证、残留验证和不可变归档，并为每次动作保存可核验的哈希链审计证据。

状态流程为：`draft` 草稿 → `counted` 已清点 → `pending_review` 待审查 → `approved` 已批准 → `destroyed` 已销毁 → `verified` 已验证 → `archived` 已归档。验证失败会进入 `remediation` 待补救，退回审查会回到清点补正。归档后批次只读。

## 构建、运行与测试

```text
go test ./...
go run ./cmd/server -addr=127.0.0.1:19081
```

服务默认监听 `127.0.0.1:19081`，也可以使用 `-addr=127.0.0.1:<port>` 或 `PORT=<port>` 配置回环端口；数据默认写入 `./data`。浏览器访问 `/`，同源 JSON 接口位于 `/api/cases`。

端到端自检会创建临时数据目录并经真实 HTTP 完成全流程，随后主动退出：

```text
go run ./cmd/server -addr=127.0.0.1:19091 -selfcheck -selfcheck-timeout=20s
```
