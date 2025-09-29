# anki-vocab

一个命令行工具，可以读取指定的单词列表，自动查询并生成内容丰富的 Anki 卡片集（`.apkg` 文件），方便您直接导入 Anki 使用。

它通过集成有道词典和火山方舟大模型（支持 DeepSeek 等），自动为单词填充释义、发音、词源、助记法、例句、近义词、故事等多种信息，极大地提升了制作 Anki 词汇卡的效率和质量。

## ✨ 主要特性

- **自动化处理**：提供单词列表，一键生成完整的 Anki 卡片集。
- **内容丰富**：
    - **词典查询**：使用有道词典查询单词释义和英/美式发音。
    - **AI 赋能**：使用火山方舟大模型生成高质量的词源、助记法、多种词性、同义词、场景搭配、例句和迷你故事。
- **高度可定制**：
    - **Anki 模板**：笔记模板（Note Type）完全可定制（HTML & CSS）。
    - **AI 指令**：可通过修改 Prompt 来自定义 AI 生成内容的风格和种类。
- **缓存支持**：自动缓存已查询的单词信息，再次生成时无需重复请求，节省时间和 API 调用成本。
- **简单易用**：通过简单的命令行即可完成所有操作。

## 🚀 安装与使用

### 📦 步骤 1: 安装

确保您已安装 Go 环境 (版本 >= 1.24)，然后在终端运行以下命令：

```bash
go install github.com/lftk/anki-vocab@latest
```

### ⚙️ 步骤 2: 初始化配置

要使用 AI 功能，您需要一个配置文件来填入 API Key。运行以下命令可生成一份默认的配置文件：

```bash
anki-vocab init --dicts dicts.yaml
```

该命令会创建 `dicts.yaml` 文件。请打开此文件，将 `api_key` 字段的值替换为您自己的火山引擎 API Key。

💡 **提示**：火山方舟目前为个人开发者提供协作奖励，每日单个模型可享 50 万免费 tokens，足够满足个人日常使用。详情请参考[官方文档](https://www.volcengine.com/docs/82379/1391869)。

如果您想了解所有可配置的选项，可以查阅项目中的 [`dicts.yaml.example`](dicts.yaml.example) 文件。

### ✍️ 步骤 3: 准备单词列表

您需要创建一个 `.txt` 格式的单词列表文件，每行一个单词。

例如，创建一个名为 `words.txt` 的文件：

```txt
// 这是一个单词表示范例，支持注释和子牌组
hello
world
## fruits // 使用 ## 创建子牌组
apple
banana
orange
```

### ⚡️ 步骤 4: 运行生成命令

打开终端，运行 `generate` 命令，并指定单词列表文件：

```bash
anki-vocab generate --name "我的词汇本" words.txt
```

命令执行完毕后，当前目录下会生成一个名为 `我的词汇本.apkg` 的文件，直接双击即可导入 Anki 客户端。

### 📋 命令行参数说明

`generate` 命令的完整参数如下：

- `--name` (必需): 指定生成的 Anki 卡片集的基础名称。
- `--output`, `-o`: 输出的 `.apkg` 文件路径。默认为 `<name>.apkg`。
- `--dicts`: 配置文件路径。默认为 `./dicts.yaml`。
- `--notetype`: 自定义笔记模板的目录路径。默认为程序内置模板。
- `--cache-dir`: 缓存目录路径。默认为用户系统缓存目录下的 `anki-vocab` 文件夹。
- `--no-cache`: 禁用缓存。
- `--verbose`, `-v`: 启用详细输出模式，会打印正在处理的每个单词。
- `wordlist_file` (位置参数, 必需): 指定输入的单词列表 `.txt` 文件路径。

## 🎨 高级自定义

本工具的核心设计思想是高度可定制。您可以从数据源（词典）、数据处理（字段模板）到最终呈现（卡片模板）进行全方位的自定义。

### 🧠 智能查询（懒加载）

这是一个核心的性能优化特性。在查询任何词典之前，程序会首先静态分析所有字段模板（[`fields/*.tmpl`](notetype/fields)）的内容，并确定模板到底需要哪些数据源。

- **按需查询**：只有当模板中明确使用了某个词典的数据（例如，出现了 `.volcengine`），程序才会去调用该词典的 API。如果所有模板都没有用到 `volcengine`，那么相关的网络请求就根本不会发生。
- **按需生成发音**：同理，只有当卡片模板中使用了 `youdao_us_pronunciation` 或 `youdao_uk_pronunciation` 字段时，程序才会去下载对应的音频文件。

这个机制可以极大地节省 API 调用次数（特别是对于付费的 AI 服务）和处理时间。

### 🌊 自定义流程概述

数据处理的流水线如下：

1.  **词典 (Dict)**: 程序根据模板所需，按需查询词典，每个词典（如 Youdao、Volcengine）会返回一个 JSON 格式的数据。这是所有信息的源头。
2.  **字段模板 (Field Template)**: 每个字段（如 `definitions`, `etymologies`）都是一个独立的 Go 模板。它负责从词典返回的 JSON 中提取需要的数据，并将其处理成最终的 HTML 片段。
3.  **卡片模板 (Card Template)**: 卡片的正面（[`front.html`](notetype/templates/Card%201/front.html)）和背面（[`back.html`](notetype/templates/Card%201/back.html)）模板，负责将多个字段的 HTML 片段组合起来，构成最终的卡片样式。其中 `word` 是一个固定的特殊字段，代表当前正在处理的单词。

### 🧩 字段模板详解 ([`fields/*.tmpl`](notetype/fields))

字段模板是自定义的核心。它使用 Go Template 语法，并内置了强大的函数来处理数据。

#### 从 JSON 中取值

您可以在字段模板中，通过 `.` 来访问词典返回的 JSON 数据的根节点。程序会将所有词典的查询结果合并到一个 JSON 对象中，对象的键是词典的名称（如 `youdao`, `volcengine`）。

为了方便您了解每个词典具体返回了哪些字段，我在 [`docs/dicts/`](docs/dicts/) 目录下提供了 [`youdao.json`](docs/dicts/youdao.json) 和 [`volcengine.json`](docs/dicts/volcengine.json) 作为数据结构样例。在自定义字段时，您可以随时查阅它们。

例如，[`etymologies.tmpl`](notetype/fields/etymologies.tmpl) 字段模板的内容如下，它从有道词典返回的数据中提取了中文词源信息，并只显示前三条：

```go-template
{{range .youdao.etym.etyms.zh | limit 3}} 
<div class="etymology"> 
    <span class="etym-word">{{.word}}</span>
    <span class="value">{{.value}}</span> 
</div>
{{end}}
```

#### 内置模板函数

为了方便处理数据，字段模板中内置了以下几个实用的函数，建议使用管道符（`|`）风格来调用：

1.  **`join`**: 将一个数组（slice）用指定的分隔符连接成一个字符串。
    *   **用法**: `{{ .数组 | join "分隔符" }}`
    *   **示例**: 假设 JSON 中有 `{"synonyms": ["a", "b", "c"]}`，以下模板：
        ```go-template
        {{ .volcengine.synonyms | join ", " }}
        ```
        会输出：`a, b, c`

2.  **`limit`**: 截取数组的前 N 个元素。
    *   **用法**: `{{ .数组 | limit N }}`
    *   **示例**: `etymologies.tmpl` 中的 `{{ .youdao.etym.etyms.zh | limit 3 }}` 就是一个绝佳的例子，它表示只处理最多 3 条词源记录。

3.  **`highlight_word`**: 在一个句子中高亮显示当前单词。它会用 `<span class="highlight">...</span>` 标签包裹单词。（注意：此函数由程序在内部注入，`word` 参数是自动处理的）
    *   **用法**: `{{ .要处理的句子 | highlight_word }}`
    *   **示例**:
        ```go-template
        {{ .sentence | highlight_word }}
        ```
        如果单词是 "apple"，句子是 "An apple a day keeps the doctor away."，则输出的 HTML 会是：
        `An <span class="highlight">apple</span> a day keeps the doctor away.`

### 🔊 发音处理机制

发音功能是基于懒加载的半自动处理。如果一个词典（如 `youdao`）实现了发音接口，且卡片模板中使用了对应的发音字段，程序才会抓取音频文件，并生成 Anki 音频标签。

*   **有道词典支持**: 目前内置的有道词典支持 `us` (美式) 和 `uk` (英式) 两种发音。
*   **使用方法**: 程序会自动生成名为 `youdao_us_pronunciation` 和 `youdao_uk_pronunciation` 的字段，其内容为 `[sound:xxxx.mp3]` 格式的 Anki 标签。您只需在卡片模板（[`front.html`](notetype/templates/Card%201/front.html) 或 [`back.html`](notetype/templates/Card%201/back.html)）中直接使用它们即可。
    *   例如，在 `front.html` 中添加：
        ```html
        美式发音: {{ youdao_us_pronunciation }}
        英式发音: {{ youdao_uk_pronunciation }}
        ```

### 🧑‍💻 为开发者：实现自定义词典

如果您希望添加本项目尚未支持的词典，您可以通过修改源码、实现 `dict.Dict` 接口来贡献新的词典源。

1.  **理解核心接口 ([`internal/dict/dict.go`](internal/dict/dict.go))**:
    *   `Queryer`: 核心接口，需要实现 `Query(ctx, word)` 方法，返回一个包含单词信息的 JSON `[]byte`。
    *   `Pronouncer`: 如果词典支持发音，则需要实现 `Pronounce(ctx, word, accent, format)` 方法，返回一个包含音频数据的 `io.ReadCloser`。
    *   `dict.Dict`: 一个结构体，包含了您的 `Queryer`、`Pronouncer` 实现和 `Capabilities`（用于声明词典能力）。

2.  **实现您的词典**:
    *   在 `internal/dict/` 目录下创建一个新的包（例如 `mydict`）。
    *   在包中实现 `Queryer` 和/或 `Pronouncer` 接口。
    *   创建一个 `New(...)` 函数，返回一个 `*dict.Dict` 实例。

3.  **注册新词典 ([`internal/registry/registry.go`](internal/registry/registry.go))**:
    *   将您的词典配置结构体添加到 `registry.config` 中。
    *   在 `registry.dicts` 这个 map 中，添加一个新条目，将词典名称（如 `"mydict"`）映射到您的 `New` 函数。

完成以上步骤后，重新编译，您的新词典就可以在 `dicts.yaml` 中配置和使用了。

## 📜 许可证

本项目基于 [GNU AGPLv3](LICENSE) 授权。
