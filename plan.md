这是一个基于 go 语言实现的完全兼容 OpenAI API 的 AI gateway。

使用 go http 标准库实现，提供完善的 hook system，可以对 AI 调用请求进行调用前和调用后处理，并可以在此基础上实现：

- OpenAI API Key 校验和替换
- token usage 统计
- Open Telemetry 集成
- LLM Provider、Models 和 API Key 管理
- 模型动态路由
- Model Name 重写替换

该项目既可以作为库被第三方服务引用，也可以作为独立的服务运行，目前我们只需要考虑作为库的方式引用