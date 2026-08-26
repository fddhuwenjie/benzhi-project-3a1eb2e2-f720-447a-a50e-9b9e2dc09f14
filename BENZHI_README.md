# BENZHI_README

## 项目说明
- 项目：benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14
- 项目用途：实验室受控材料退役销毁见证台提供从退役建档、双人清点、风险审查、现场销毁、残留验证到不可变归档的完整浏览器流程，并以本地快照和哈希链审计证据支撑并发与合规核验。
- Go 工具链：`golang:1.23`
- 前端工具链：原生 HTML、CSS 和 JavaScript，由 Go 服务直接提供

## 项目描述
- 项目名称：实验室受控材料退役销毁见证台
- 项目介绍：面向实验室安全管理员与现场见证人员的单流程浏览器应用，用于将受控实验材料从退役建档、双人清点、风险审查、现场销毁、残留验证推进到不可变归档，并保留可核验的状态与审计证据。
- 项目概述：面向实验室安全管理员与现场见证人员的单流程浏览器应用，用于将受控实验材料从退役建档、双人清点、风险审查、现场销毁、残留验证推进到不可变归档，并保留可核验的状态与审计证据。
- 核心工作流：受控材料退役批次从草稿建档开始，依次完成双人清点、风险审查与批准、现场销毁见证、残留验证，最终生成封存摘要并进入不可修改的已归档状态；审查退回或验证失败时只能沿本流程补正后重新提交。
- 对外接口：Go 服务提供原生 HTML、CSS 和 JavaScript 的浏览器工作台，通过同源 JSON 接口完成批次列表、详情、表单操作、状态提示和审计时间线；不引入 Node 构建链。

## 标准构建、运行和测试命令
进入容器后执行：
cd '/app' && GOTOOLCHAIN=local go build ./...

cd /app && GOTOOLCHAIN=local go run ./cmd/server -addr=127.0.0.1:19091 -selfcheck -selfcheck-timeout=20s

cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
chmod +x build_benzhi_docker.sh

./build_benzhi_docker.sh benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14-amd64 linux/amd64

./build_benzhi_docker.sh benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14-arm64 linux/arm64

docker run -it benzhi-project-3a1eb2e2-f720-447a-a50e-9b9e2dc09f14-amd64:latest

## 题目验证命令
1. 预期退出码 0：`go test ./...`
2. 预期退出码 0：`go run ./cmd/server -addr=127.0.0.1:19091 -selfcheck -selfcheck-timeout=20s`
