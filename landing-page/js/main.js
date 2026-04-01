/**
 * Clipal Landing Page — js/main.js
 * Modules: i18n · doc-links · terminal · scroll-reveal · os-detect · platform-tabs · install-tabs · clipboard
 */
'use strict';

/* ============================================================
   1. DOC LINKS (language-aware)
   ============================================================ */
const DOC_LINKS = {
  en: {
    docs: 'https://github.com/lansespirit/Clipal/tree/main/docs/en',
    'getting-started': 'https://github.com/lansespirit/Clipal/blob/main/docs/en/getting-started.md',
    config: 'https://github.com/lansespirit/Clipal/blob/main/docs/en/config-reference.md',
    webui: 'https://github.com/lansespirit/Clipal/blob/main/docs/en/web-ui.md',
  },
  zh: {
    docs: 'https://github.com/lansespirit/Clipal/tree/main/docs/zh-CN',
    'getting-started': 'https://github.com/lansespirit/Clipal/blob/main/docs/zh-CN/getting-started.md',
    config: 'https://github.com/lansespirit/Clipal/blob/main/docs/zh-CN/config-reference.md',
    webui: 'https://github.com/lansespirit/Clipal/blob/main/docs/zh-CN/web-ui.md',
  },
};

function updateDocLinks(lang) {
  const links = DOC_LINKS[lang] || DOC_LINKS.en;
  document.querySelectorAll('[data-doclink]').forEach(el => {
    const key = el.dataset.doclink;
    if (links[key]) el.href = links[key];
  });
}

/* ============================================================
   2. TRANSLATIONS (i18n)
   ============================================================ */
const TRANSLATIONS = {
  en: {
    'nav.features': 'Features',
    'nav.install': 'Quick Start',
    'nav.howitworks': 'How It Works',
    'nav.docs': 'Docs',
    'nav.github': 'GitHub',

    'hero.eyebrow': 'LLM Proxy · No Docker · Open Source',
    'hero.line1': 'Universal CLI Takeover.',
    'hero.line2': 'Claude Code, Codex & Gemini.',
    'hero.line3': 'One Binary to Rule Them All.',
    'hero.sub': 'The professional LLM infrastructure for Claude Code, Codex, and Gemini CLI. Universal gateway for any tool supporting custom BASE_URL. No Docker, no Python, just performance.',
    'hero.cta.download': 'Download Clipal',
    'hero.cta.github': 'Star on GitHub',

    'ecosystem.title': 'Claude Code, Codex & Gemini. Natively.',
    'ecosystem.sub': 'Directly patches ~/.claude/settings.json & ~/.codex/config.toml. Clipal also serves as a unified proxy for Aider, Continue, Jan, Cherry Studio, Chatbox, and any app supporting custom API endpoints.',
    'ecosystem.protocols': 'Core Protocols',
    'ecosystem.cli': 'CLI Tools & AI LLM Applications',

    'features.eyebrow': 'Core Advantages',
    'features.title': 'Built Different. Works Everywhere.',
    'features.sub': 'Three things that make Clipal the only LLM proxy worth running.',
    'features.takeover.title': 'Universal LLM CLI Takeover',
    'features.takeover.desc': 'Stop hunting for hidden config files. Clipal natively patches ~/.claude/settings.json and standard configuration files, transforming into a zero-configuration local reverse proxy for Claude Code, AI CLI tools, and Desktop apps without env-var pollution.',
    'features.failover.title': 'Smart 429 Bypass & Failover Routing',
    'features.failover.desc': 'Conquer strict 3rd-party API rate limits. Clipal manages shared token pools, instantly detecting 429 errors to gracefully rotate keys or failover to backup providers—all while preserving your contextual session state.',
    'features.failover.stat1': 'Dropped Requests',
    'features.failover.stat2': 'Keys per Provider',
    'features.failover.stat3': 'Unified Failover',
    'features.simple.title': 'Lightweight Single-Binary LLM Gateway',
    'features.simple.desc': 'Unlike heavy Python proxies (like LiteLLM) or complex Docker deployments, Clipal is a high-performance, zero-dependency API Gateway compiled in Go for macOS, Linux, and Windows. Download. Executable. Proxy active.',
    'features.simple.other': '❌ Other Solutions',
    'features.simple.vs': 'vs',

    'hiw.eyebrow': 'Behind the Magic',
    'hiw.title': 'The Clipal Blueprint',
    'hiw.sub': 'How we engineer stability: zero config loss, zero auth loops, and unified token pools.',
    'hiw.client.cli.title': 'AI CLI Tools',
    'hiw.client.cli': 'Claude Code, Codex, Gemini CLI, OpenCode. Fragile local base_url configs gracefully patched.',
    'hiw.client.apps.title': 'AI Desktop Apps',
    'hiw.client.apps': 'Cherry Studio, OpenClaw, ChatWise. Standardized flows behind a single gateway.',
    'hiw.core1.tag': 'AST-Merge Engine',
    'hiw.core1.title': 'Non-Destructive Takeover',
    'hiw.core1.desc': 'Instead of overwriting settings.json, we parse the JSON AST and surgically patch ANTHROPIC_BASE_URL. Your enabledPlugins and custom env variables remain 100% untouched.',
    'hiw.core2.tag': 'Protocol Normalizer',
    'hiw.core2.title': 'Socket-Level Header Scrubbing',
    'hiw.core2.desc': 'Tired of Login Loops? Clipal actively scrubs conflicting Host headers and injects accurate API tokens at the proxy layer, enabling flawless AWS Bedrock / Google Vertex connections.',
    'hiw.core3.tag': 'Routing Multi-Pool',
    'hiw.core3.title': 'Intelligent 429 & Sticky Sessions',
    'hiw.core3.desc': 'Built for 3rd-party APIs. When a single Key hits 429 rate limits, Clipal instantly diverts to the next available Key within the same Provider, while enforcing Context Sticky Sessions so conversations never lose state.',
    'hiw.core4.tag': 'Hierarchical Failover',
    'hiw.core4.title': 'Multi-Provider Fallback',
    'hiw.core4.desc': 'Total endpoint outage? If a Primary Provider exhausts all keys or goes offline, priority-based hierarchical routing seamlessly cascades traffic to your Fallback Provider. Zero downtime.',
    'hiw.out.claude': '100% Protocol match. Zero authentication loops.',
    'hiw.out.openai': 'High-performance Codex proxying, full OpenAI spec.',
    'hiw.out.gemini': 'Rotating quotas softly, bypassing RPS limits transparently.',

    'why.eyebrow': 'Competitive Advantage',
    'why.title': 'Why Developers Switch to Clipal',
    'why.sub': 'Beyond simple proxying — Clipal is the infrastructure that cc-switch and ccNexus users have been waiting for.',
    'why.highlights.title': 'Clipal Signature Capabilities',
    'why.claude.title': 'Claude Code Proxy: Native & Powerful',
    'why.claude.desc': 'No more manual ANTHROPIC_BASE_URL exports. Clipal natively handles Claude Code session hijacking and header scrubbing for Bedrock/Vertex AI.',
    'why.gemini.title': 'Codex & Gemini: Quota Master',
    'why.gemini.desc': 'Automatically rotate multiple keys to bypass RPS limits. Fallback to Flash when Pro hits the ceiling. Keep your coding flow uninterrupted.',

    'comp.feature': 'Feature / Tool',
    'comp.binary': 'Binary Only (No Docker)',
    'comp.takeover': 'Claude / Codex / Gemini Takeover',
    'comp.pools': 'API Provider Multi-key Pools',
    'comp.logs': 'Universal Base-url Support',
    'comp.non-destructive': 'Non-destructive Config Merge',
    'comp.auth-emulation': 'Protocol Auth Emulation',
    'comp.diagnostics': 'Real-time Flow Monitoring',
    'comp.val.ok': '✓',
    'comp.val.err': '✗',
    'comp.val.limited': 'Limited',
    'comp.val.binary.ccs': '✗ (Script)',
    'comp.val.unified': 'Unified',
    'comp.val.bulletproof': 'Bulletproof',
    'comp.val.metrics': '✓ (w/ Metrics)',
    'comp.val.overwrites': '✗ (Overwrites)',
    'comp.val.safeguarded': '✓ (Safeguarded)',
    'comp.val.no-loops': '✓ (No Loops)',
    'comp.val.blackbox': '✗ (Blackbox)',
    'comp.val.one-click': '✓ (One-click)',

    'faq.eyebrow': 'F.A.Q.',
    'faq.title': 'Frequently Asked Questions for Clipal',
    'faq.q1': 'Is Clipal a local SaaS or cloud proxy?',
    'faq.a1': 'Neither. Clipal is a standalone local binary. Your API keys are stored in an encrypted config.json on your own machine. Zero cloud storage, zero data leakage.',
    'faq.q2': 'How is Clipal different from cc-switch?',
    'faq.a2': 'While cc-switch manages multiple configs, Clipal is a high-performance gateway that supports complex multi-key token pools, deep config injection for clauderc, and real-time observability.',
    'faq.q3': 'Does Clipal solve "Invalid API Key" for Bedrock?',
    'faq.a3': 'Yes. Most proxies break when talking to Bedrock/Vertex via Claude Code due to header strictness. Clipal provides native protocol emulation to fix these issues at the socket level.',
    'faq.q4': 'Is there any performance overhead?',
    'faq.a4': 'Negligible. Written in Go, Clipal processes requests in microseconds. It consumes less than 50MB of RAM, far more efficient than Node.js or Python-based proxies.',
    'faq.q5': 'How does Clipal protect my existing configuration?',
    'faq.a5': 'Unlike other tools that overwrite your settings, Clipal uses a non-destructive merging strategy. We only update necessary fields and preserve your enabledPlugins, custom environment variables, and manual tweaks.',
    'faq.q6': "I'm tired of the 'Login Loop' in Claude Code. Can Clipal fix this?",
    'faq.a6': "Yes. Most Login Loops are caused by incorrect header handling when proxying. Clipal's protocol emulation layer ensures headers are properly scrubbed and injected, keeping your session stable without repeated authentication prompts.",
    'faq.q7': 'How do I know if the proxy is actually working?',
    'faq.a7': 'Clipal provides a real-time observability dashboard. You can see the full request flow, response times, and error codes in one click. If a request fails, we tell you exactly why.',

    'install.eyebrow': 'Quick Start',
    'install.title': 'Up and Running Clipal in Minutes',
    'install.sub': 'Choose your preferred installation method.',
    'install.tab.ai': 'Let AI Install for Me',
    'install.tab.manual': 'Manual Install',
    'install.ai.hint': 'Copy this prompt and paste it into Claude Code, Aider, or any AI coding assistant — it will handle the entire installation for you:',
    'install.copy': 'Copy',
    'install.copied': 'Copied!',
    'install.step1.title': 'Download the Binary',
    'install.step1.btn': 'Download Binary',
    'install.step2.title': 'Move to PATH & Start',
    'install.step3.title': 'Open Web UI & Configure',
    'install.step3.desc': 'Visit http://127.0.0.1:3333 to add your API keys, apply CLI Takeover, and manage providers.',

    'steps.eyebrow': 'How It Works',
    'steps.title': 'One Gateway, Every AI Tool',
    'steps.1.title': 'Download & Install',
    'steps.1.caption': 'One binary. No dependencies. Works on macOS, Linux, and Windows.',
    'steps.2.title': 'Start the Service',
    'steps.2.caption': 'Runs silently in the background. Persists across reboots.',
    'steps.3.title': 'One-Click CLI Takeover',
    'steps.3.caption': 'Open the dashboard. One click. All your AI tools now route through Clipal.',

    'download.eyebrow': 'Get Clipal',
    'download.title': 'Free & Open Source. Always.',
    'download.panel.title': 'Download for Your Platform',
    'download.manual': 'Or download directly:',
    'download.btn': 'Download Binary',
    'download.version': 'Latest stable:',
    'download.all': 'All releases →',
    'download.security': '100% local. Your keys never leave your machine. MIT License.',
    'platform.mac-arm': 'macOS (Apple Silicon)',
    'platform.mac-intel': 'macOS (Intel)',
    'platform.linux-x64': 'Linux x86_64',
    'platform.linux-arm': 'Linux ARM64',
    'platform.windows': 'Windows',

    'community.title': 'Join the Community',
    'community.sub': 'Get setup help, share configs, and discuss new providers with other developers.',
    'community.qr.label': 'Scan to join WeChat group',

    'footer.tagline': 'Your Local LLM Gateway',
    'footer.paiart': 'Part of the PAIART ecosystem — Empowering Personal AI Communities.',
    'footer.product': 'Product',
    'footer.resources': 'Resources',
    'footer.community': 'Community',
    'footer.download': 'Download',
    'footer.gettingstarted': 'Getting Started',
    'footer.config': 'Config Reference',
    'footer.releases': 'Release Notes',
    'footer.license': 'MIT License',
    'footer.copyright': '© 2025 Clipal Contributors. MIT License.',
    'footer.security': '100% local. Your keys never leave your machine.',

    'install.ai.prompt': `Please help me install and start Clipal. Project: https://github.com/lansespirit/Clipal

Please detect my current OS and architecture, check the project's Releases and docs, and complete the download, installation, and startup for me. Then confirm that I can open the Web UI successfully. Use these official links when needed:
- Releases: https://github.com/lansespirit/Clipal/releases
- Getting Started: https://github.com/lansespirit/Clipal/blob/main/docs/en/getting-started.md
- Web UI Guide: https://github.com/lansespirit/Clipal/blob/main/docs/en/web-ui.md

After that, guide me through using the Web UI to enable CLI takeover and add my first provider.`,
  },

  zh: {
    'nav.features': '功能特性',
    'nav.install': '快速开始',
    'nav.howitworks': '工作原理',
    'nav.docs': '文档',
    'nav.github': 'GitHub',

    'hero.eyebrow': 'LLM 代理 · 无 Docker · 开源免费',
    'hero.line1': '一键接管 AI CLI.',
    'hero.line2': '掌控 Claude Code, Codex & Gemini.',
    'hero.line3': '一个二进制文件，接管所有。',
    'hero.sub': '专为 Claude Code, Codex 和 Gemini CLI 打造的专业级 LLM 基础设施。支持任何自定义 BASE_URL 工具的通用网关。无需 Docker, 无需 Python，追求极致性能。',
    'hero.cta.download': '下载 Clipal',
    'hero.cta.github': 'Star on GitHub',

    'ecosystem.title': '原生支持 Claude Code, Codex & Gemini',
    'ecosystem.sub': '直接补丁 ~/.claude/settings.json 与 ~/.codex/config.toml。Clipal 也是 Aider, Continue, Cherry studio, Jan, Chatbox 及任何支持自定义 API 端点应用的统一代理。',
    'ecosystem.protocols': '核心支持协议',
    'ecosystem.cli': 'CLI 工具与 AI LLM 应用生态',

    'features.eyebrow': '核心优势',
    'features.title': '与众不同，适用一切。',
    'features.sub': '三个让 Clipal 成为唯一值得运行的 LLM 代理的理由。',
    'features.takeover.title': '毫秒级 LLM 命令行接管',
    'features.takeover.desc': '告别繁琐的手动改配置。Clipal 能够自动注入 ~/.claude/settings.json，化身 Claude Code 和各类 AI 工具的专属本地反向代理，全程零环境变量污染。',
    'features.failover.title': '智能 429 过载分流与多级兜底',
    'features.failover.desc': '彻底解决 3rd-party 中转 API 频繁限流痛点。Clipal 支持统一 Token 池管理，毫秒级感知 429 报错并自动轮换多 Key，甚至在节点宕机时将流量无损转移至备用路线。',
    'features.failover.stat1': '配置与报错拦截',
    'features.failover.stat2': '密钥池无限轮换',
    'features.failover.stat3': '流量层级兜底',
    'features.simple.title': '零依赖的极速单体本地网关',
    'features.simple.desc': '不同于臃肿的 Python 依赖 (如 LiteLLM) 或是复杂的 Docker Compose 编排，Clipal 是完全基于 Go 开发的高并发独立二进制网关。下载即用，告别环境红字报错。',
    'features.simple.other': '❌ 其他方案',
    'features.simple.vs': 'vs',

    'hiw.eyebrow': '幕后技术',
    'hiw.title': 'Clipal 技术架构蓝图',
    'hiw.sub': '我们如何从底层保障核心稳定性：零配额耗尽、零登录死循环、零配置覆盖。',
    'hiw.client.cli.title': '命令行 AI 助手',
    'hiw.client.cli': 'Claude Code, Codex, Gemini CLI, OpenCode 等。频繁覆写的配置受到完美保护。',
    'hiw.client.apps.title': 'AI 桌面端应用',
    'hiw.client.apps': 'Cherry Studio, OpenClaw, ChatWise 等。所有流量统一收口于一个本地稳定网关。',
    'hiw.core1.tag': 'AST 语法树精准合并',
    'hiw.core1.title': '非破坏性配置接管',
    'hiw.core1.desc': '不同于粗暴覆盖 settings.json，我们解析 JSON AST 仅精准注入 ANTHROPIC_BASE_URL。你的 enabledPlugins 及自定义特殊变量受到 100% 保护，并提供内置快照备份。',
    'hiw.core2.tag': '协议级防篡改网关',
    'hiw.core2.title': 'Socket 级 Header 清洗',
    'hiw.core2.desc': '告别无限次要求 /login 的噩梦。Clipal 在代理网关层直接重写 Auth 请求并强制涤除导致握手失败的混乱 Host Header，使 AWS Bedrock / Vertex AI 顺畅直连。',
    'hiw.core3.tag': '多密钥调度池与重试',
    'hiw.core3.title': '429 智能分流与会话粘性',
    'hiw.core3.desc': '专为第三方 API 设计。单渠道配额耗尽或触发 429 限流？网关会在同一 Provider 池内瞬间无感切换至下一个 Key；智能会话粘性逻辑保障每次上下文路由都尽可能沿用前代环境，实现零断点对话保护。',
    'hiw.core4.tag': '多级 Provider 路由',
    'hiw.core4.title': '全局层级兜底 (Failover)',
    'hiw.core4.desc': '不怕单点全垒崩溃。如果你的主力 Provider 彻底宕机或者所有 Key 均被耗尽，优先级路由会把整个请求安全转移至备用的、不同基座的 Provider 上。',
    'hiw.out.claude': '1:1 反向模拟协议，彻底粉碎 401 拒权死循环。',
    'hiw.out.openai': '兼容所有遵循 OpenAI 标准的第三方 API 与中转节点。',
    'hiw.out.gemini': '对外部频控限制视若无影，动态池平滑引流不掉线。',

    'why.eyebrow': '差异化竞争优势',
    'why.title': '为什么资深开发者选择 Clipal',
    'why.sub': '不仅仅是一个代理。Clipal 是 cc-switch 和 ccNexus 用户期待已久的“精装修”基础设施。',
    'why.highlights.title': '针对主流工具的生态优化',
    'why.claude.title': 'Claude Code: 协议抹平专家',
    'why.claude.desc': '告别手动 export 环境变量。Clipal 自动注入 ANTHROPIC_BASE_URL 并处理 Bedrock/Vertex 严格的协议头校验。',
    'why.gemini.title': 'Codex & Gemini: 配额/中转首选',
    'why.gemini.desc': '自动轮询多密钥以突破 RPS 频率检测。Pro 配额耗尽时秒切 Flash 镜像模型，确保编码工作流永不中断。',

    'comp.feature': '核心对比指标',
    'comp.binary': '单一二进制 (无 Docker)',
    'comp.takeover': 'Claude / Gemini 配置接管',
    'comp.pools': 'Codex 多密钥 Token 池',
    'comp.logs': '通用 Base-url 应用支持',
    'comp.non-destructive': '非破坏性配置接管',
    'comp.auth-emulation': '原生协议认证仿真',
    'comp.diagnostics': '实时流量观测',
    'comp.val.ok': '✓',
    'comp.val.err': '✗',
    'comp.val.limited': '有限支持',
    'comp.val.binary.ccs': '✗ (脚本)',
    'comp.val.unified': '全生态接管',
    'comp.val.bulletproof': '极稳密钥池',
    'comp.val.metrics': '✓ (附带观测)',
    'comp.val.overwrites': '✗ (暴力覆盖)',
    'comp.val.safeguarded': '✓ (JSON 合并)',
    'comp.val.no-loops': '✓ (无死循环)',
    'comp.val.blackbox': '✗ (黑盒运行)',
    'comp.val.one-click': '✓ (实时面板)',

    'faq.eyebrow': '常见问题解答',
    'faq.title': '关于 Clipal 的常见问题',
    'faq.q1': 'Clipal 是本地工具还是云端代理？',
    'faq.a1': 'Clipal 是纯本地运行的二进制文件。你的 API Key 存储在本地被简单加密的 config.json 中，绝不经过任何云端。',
    'faq.q2': 'Clipal 与 cc-switch 有什么区别？',
    'faq.a2': 'cc-switch 主要负责多配置切换，而 Clipal 是高性能网关，不仅通过非破坏性合并保护你的 settings.json，还支持多密钥轮询和实时请求观测。',
    'faq.q3': 'Clipal 能解决 Bedrock 的 "Invalid Header" 报错吗？',
    'faq.a3': '可以。绝大多数代理在处理 Claude Code 连接 Bedrock/Vertex 时会因为 Header 校验失败。Clipal 提供原生协议仿真，从 Socket 层级修复此问题。',
    'faq.q4': '运行性能如何？',
    'faq.a4': '极度轻量。使用 Go 开发，内存占用低于 50MB，内部协议转发延迟低于 1 毫秒。',
    'faq.q5': 'Clipal 如何保护我已有的配置文件？',
    'faq.a5': '不同于其他工具直接覆盖整个 settings.json，Clipal 采用非破坏性策略。我们仅更新必要的 Provider 字段，并完整保留你的 enabledPlugins 和自定义环境变量。',
    'faq.q6': '我厌倦了 Claude Code 的登录循环，Clipal 能解决吗？',
    'faq.a6': '是的。登录循环通常是由代理时 Header 处理不当引起的。Clipal 的协议仿真层确保护理所有 Header 注入，让你的会话保持稳定，告别重复认证。',
    'faq.q7': '我如何知道代理是否真的在工作？',
    'faq.a7': 'Clipal 提供实时观测面板。点击即可查看完整的请求流、响应时间和错误代码。如果请求失败，我们会明确告知是网络超时还是 API Key 无效。',

    'install.eyebrow': '快速开始',
    'install.title': '几分钟内启动 Clipal',
    'install.sub': '选择你偏好的安装方式。',
    'install.tab.ai': '让 AI 帮我安装',
    'install.tab.manual': '手动安装',
    'install.ai.hint': '复制以下提示词，粘贴到 Claude Code、Aider 或任何 AI 编程助手中 — 它将为你完成全部安装：',
    'install.copy': '复制',
    'install.copied': '已复制！',
    'install.step1.title': '下载二进制文件',
    'install.step1.btn': '下载二进制',
    'install.step2.title': '加入 PATH 并启动',
    'install.step3.title': '打开 Web UI 并配置',
    'install.step3.desc': '访问 http://127.0.0.1:3333，添加 API Key，启用 CLI 接管，管理提供商。',

    'steps.eyebrow': '工作原理',
    'steps.title': '一个网关，所有 AI 工具',
    'steps.1.title': '下载并安装',
    'steps.1.caption': '单一二进制文件，无依赖。支持 macOS、Linux 和 Windows。',
    'steps.2.title': '启动服务',
    'steps.2.caption': '静默在后台运行，开机自启。',
    'steps.3.title': '一键接管 CLI',
    'steps.3.caption': '打开管理面板，一键点击，你的所有 AI 工具立即通过 Clipal 路由。',

    'download.eyebrow': '获取 Clipal',
    'download.title': '永久免费与开源。',
    'download.panel.title': '选择你的平台下载',
    'download.manual': '或直接下载：',
    'download.btn': '下载二进制文件',
    'download.version': '最新稳定版：',
    'download.all': '查看所有版本 →',
    'download.security': '100% 本地运行，API Key 永不离开你的机器。MIT 开源协议。',
    'platform.mac-arm': 'macOS (Apple Silicon)',
    'platform.mac-intel': 'macOS (Intel)',
    'platform.linux-x64': 'Linux x86_64',
    'platform.linux-arm': 'Linux ARM64',
    'platform.windows': 'Windows',

    'community.title': '加入社区',
    'community.sub': '与其他开发者交流配置心得、获取帮助、探讨新提供商接入。',
    'community.qr.label': '扫码加入微信交流群',

    'footer.tagline': '你的本地 LLM 网关',
    'footer.paiart': '隶属 PAIART 生态 — 致力于构建个人 AI 社区。',
    'footer.product': '产品',
    'footer.resources': '资源',
    'footer.community': '社区',
    'footer.download': '下载',
    'footer.gettingstarted': '快速入门',
    'footer.config': '配置参考',
    'footer.releases': '发布日志',
    'footer.license': 'MIT 协议',
    'footer.copyright': '© 2025 Clipal 贡献者。MIT 开源协议。',
    'footer.security': '100% 本地运行。API Key 永不离开你的机器。',

    'install.ai.prompt': `请帮我安装并启动 Clipal。项目地址：https://github.com/lansespirit/Clipal

请检测我当前的操作系统和架构，查看项目的 Releases 和文档，并为我完成下载、安装和启动。然后确认我能成功打开 Web UI。需要时请使用以下官方链接：
- Releases：https://github.com/lansespirit/Clipal/releases
- 快速入门：https://github.com/lansespirit/Clipal/blob/main/docs/zh-CN/getting-started.md
- Web UI 指南：https://github.com/lansespirit/Clipal/blob/main/docs/zh-CN/web-ui.md

完成后，请引导我通过 Web UI 启用 CLI 接管并添加我的第一个提供商。`,
  },
};

let currentLang = localStorage.getItem('clipal-lang') || 'en';

function applyTranslations(lang) {
  currentLang = lang;
  localStorage.setItem('clipal-lang', lang);
  document.documentElement.lang = lang;
  const t = TRANSLATIONS[lang] || TRANSLATIONS.en;

  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.dataset.i18n;
    if (t[key] !== undefined) el.textContent = t[key];
  });

  // Update AI install prompt
  const promptEl = document.getElementById('ai-prompt-content');
  if (promptEl) promptEl.textContent = t['install.ai.prompt'] || '';

  // Update lang toggle styling
  const btn = document.getElementById('lang-toggle');
  if (btn) {
    btn.setAttribute('aria-label', lang === 'en' ? 'Switch to Chinese' : '切换到英文');
    btn.querySelector('.lang-en').style.color = lang === 'en' ? 'var(--gold)' : '';
    btn.querySelector('.lang-en').style.fontWeight = lang === 'en' ? '700' : '400';
    btn.querySelector('.lang-zh').style.color = lang === 'zh' ? 'var(--gold)' : '';
    btn.querySelector('.lang-zh').style.fontWeight = lang === 'zh' ? '700' : '400';
  }

  updateDocLinks(lang);
  document.body.classList.remove('is-loading');
}

function initI18n() {
  applyTranslations(currentLang);
  const btn = document.getElementById('lang-toggle');
  if (btn) btn.addEventListener('click', () => applyTranslations(currentLang === 'en' ? 'zh' : 'en'));
}

/* ============================================================
   3. TERMINAL ANIMATION
   ============================================================ */
const TERMINAL_SCRIPT = [
  { type: 'prompt', text: '$ clipal service start', delay: 400, speed: 42 },
  {
    type: 'ascii', text:
      `  ██████╗██╗     ██╗██████╗  █████╗ ██╗
 ██╔════╝██║     ██║██╔══██╗██╔══██╗██║
 ██║     ██║     ██║██████╔╝███████║██║
 ██║     ██║     ██║██╔═══╝ ██╔══██║██║
 ╚██████╗███████╗██║██║     ██║  ██║███████╗
  ╚═════╝╚══════╝╚═╝╚═╝     ╚═╝  ╚═╝╚══════╝`, delay: 200, instant: true
  },
  { type: 'blank', text: '', delay: 100 },
  { type: 'ok', text: '✓ Clipal v0.11.6 running on http://127.0.0.1:3333', delay: 180, speed: 18 },
  { type: 'ok', text: '✓ Providers loaded: Anthropic, OpenAI (3), Gemini (2)', delay: 160, speed: 18 },
  { type: 'ok', text: '✓ Key pool: 5 keys active across 2 providers', delay: 500, speed: 18 },
  { type: 'blank', text: '', delay: 200 },
  { type: 'log', text: '→ [claude code]  POST /clipal  →  claude-4-6-sonnet  ✓  288ms', delay: 700, speed: 22 },
  { type: 'log', text: '→ [opencode]        POST /clipal  →  gemini-3.1-flash-lite       ✓  341ms', delay: 450, speed: 22 },
  { type: 'log-fail', text: '→ [codex]  POST /clipal  →  gpt-5.4 [QUOTA]   ✗', delay: 900, speed: 22 },
  { type: 'log-fo', text: '                           →  OpenAI  [FAILOVER]  ✓  519ms', delay: 280, speed: 22 },
];

function escHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

function buildLine(type, text) {
  const el = document.createElement('div');
  el.style.minHeight = '1.6em';
  switch (type) {
    case 'prompt': el.innerHTML = `<span class="t-prompt">${escHtml(text)}</span>`; break;
    case 'ascii': el.innerHTML = `<span class="t-ascii">${escHtml(text)}</span>`; break;
    case 'blank': el.innerHTML = ' '; break;
    case 'ok': el.innerHTML = `<span class="t-ok">${escHtml(text)}</span>`; break;
    case 'log': {
      const arr = text.split('✓');
      el.innerHTML = `<span class="t-arrow">→</span><span class="t-dim"> ${escHtml(arr[0].replace('→', '').trim())} </span>`
        + (arr[1] ? `<span class="t-ok">✓${escHtml(arr[1])}</span>` : '');
      break;
    }
    case 'log-fail':
      el.innerHTML = `<span class="t-arrow">→</span><span class="t-dim"> ${escHtml(text.replace('→', '').replace('✗', '').trim())} </span><span class="t-err">✗</span>`;
      break;
    case 'log-fo':
      el.innerHTML = `<span class="t-dim">                              </span><span class="t-arrow">→</span><span class="t-dim"> OpenAI </span><span class="t-tag">[FAILOVER]</span><span class="t-ok"> ✓  519ms</span>`;
      break;
    default: el.textContent = text;
  }
  return el;
}

function sleep(ms) { return new Promise(r => setTimeout(r, ms)); }

async function typewriteLine(el, outputEl, speed) {
  // Extract each span's text, typewrite them in sequence
  const fullHTML = el.innerHTML;
  const tmp = document.createElement('div');
  tmp.innerHTML = fullHTML;
  el.innerHTML = '';
  outputEl.appendChild(el);

  for (const node of [...tmp.childNodes]) {
    if (node.nodeType === Node.TEXT_NODE) {
      const tn = document.createTextNode('');
      el.appendChild(tn);
      for (let i = 0; i < node.textContent.length; i++) {
        tn.textContent = node.textContent.slice(0, i + 1);
        await sleep(speed + Math.random() * 6 - 3);
      }
    } else {
      const span = node.cloneNode(false);
      el.appendChild(span);
      const full = node.textContent;
      for (let i = 0; i < full.length; i++) {
        span.textContent = full.slice(0, i + 1);
        await sleep(speed + Math.random() * 6 - 3);
      }
    }
  }
}

async function runTerminal(outputEl) {
  for (const line of TERMINAL_SCRIPT) {
    await sleep(line.delay || 150);
    const el = buildLine(line.type, line.text);

    if (line.instant || line.type === 'blank' || line.type === 'ascii') {
      outputEl.appendChild(el);
    } else {
      await typewriteLine(el, outputEl, line.speed || 28);
    }
    outputEl.scrollTop = outputEl.scrollHeight;
  }

  const cursor = document.createElement('span');
  cursor.className = 't-cursor';
  cursor.setAttribute('aria-hidden', 'true');
  outputEl.appendChild(cursor);

  await sleep(5500);
  outputEl.innerHTML = '';
  runTerminal(outputEl);
}

function initTerminal() {
  const outputEl = document.getElementById('terminal-output');
  if (!outputEl) return;
  setTimeout(() => runTerminal(outputEl), 900);
}

/* ============================================================
   4. SCROLL REVEAL
   ============================================================ */
function initScrollReveal() {
  const observer = new IntersectionObserver(entries => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const siblings = [...entry.target.parentElement.querySelectorAll('.reveal')];
        const idx = siblings.indexOf(entry.target);
        setTimeout(() => entry.target.classList.add('visible'), idx * 80);
        observer.unobserve(entry.target);
      }
    });
  }, { threshold: 0.08, rootMargin: '0px 0px -32px 0px' });

  document.querySelectorAll('.reveal').forEach(el => observer.observe(el));
}

/* ============================================================
   5. OS DETECTION & DOWNLOAD LINKS
   ============================================================ */
const RELEASE_BASE = 'https://github.com/lansespirit/Clipal/releases/latest/download';
const PLATFORM_CONFIG = {
  'mac-arm': { file: 'clipal-darwin-arm64', badge: 'Apple Silicon' },
  'mac-intel': { file: 'clipal-darwin-amd64', badge: 'Intel Mac' },
  'linux-x64': { file: 'clipal-linux-amd64', badge: 'Linux x86_64' },
  'linux-arm': { file: 'clipal-linux-arm64', badge: 'Linux ARM64' },
  'windows': { file: 'clipal-windows-amd64.exe', badge: 'Windows' },
};

function detectPlatform() {
  const ua = (navigator.userAgent || '').toLowerCase();
  const pl = (navigator.userAgentData?.platform || navigator.platform || '').toLowerCase();
  if (ua.includes('win')) return 'windows';
  if (ua.includes('mac') || pl.includes('mac')) {
    const canvas = document.createElement('canvas');
    const gl = canvas.getContext('webgl');
    const ext = gl && gl.getExtension('WEBGL_debug_renderer_info');
    const renderer = ext ? gl.getParameter(ext.UNMASKED_RENDERER_WEBGL) : '';
    return (renderer.toLowerCase().includes('apple') || navigator.maxTouchPoints > 0)
      ? 'mac-arm' : 'mac-intel';
  }
  if (ua.includes('linux')) return (ua.includes('arm') || ua.includes('aarch64')) ? 'linux-arm' : 'linux-x64';
  return null;
}

function applyPlatform(key) {
  const cfg = PLATFORM_CONFIG[key];
  if (!cfg) return;
  const url = `${RELEASE_BASE}/${cfg.file}`;

  // Hero CTA
  const heroCta = document.getElementById('cta-primary');
  const badge = document.getElementById('cta-platform-badge');
  if (heroCta) heroCta.href = url;
  if (badge) { badge.textContent = cfg.badge; badge.classList.add('visible'); }

  // Download section direct btn
  const dlBtn = document.getElementById('download-direct-btn');
  const dlText = document.getElementById('download-btn-text');
  if (dlBtn) dlBtn.href = url;
  if (dlText) dlText.textContent = cfg.file;

  // Manual install btn
  const manualBtn = document.getElementById('manual-dl-btn');
  const manualText = document.getElementById('manual-dl-text');
  if (manualBtn) manualBtn.href = url;
  if (manualText) manualText.textContent = cfg.file;

  // Activate matching tabs (both download section + install section)
  document.querySelectorAll('.ptab').forEach(tab => {
    const active = tab.dataset.platform === key;
    tab.classList.toggle('active', active);
    tab.setAttribute('aria-selected', String(active));
  });
}

function initOSDetection() {
  const detected = detectPlatform();
  if (detected) applyPlatform(detected);
}

/* ============================================================
   6. PLATFORM TABS (download + install sections)
   ============================================================ */
function initPlatformTabs() {
  document.querySelectorAll('.ptab').forEach(tab => {
    tab.addEventListener('click', () => applyPlatform(tab.dataset.platform));
  });
}

/* ============================================================
   7. INSTALL TABS (AI / Manual)
   ============================================================ */
function initInstallTabs() {
  const tabs = document.querySelectorAll('.itab');
  const panels = document.querySelectorAll('.itab-panel');
  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      tabs.forEach(t => { t.classList.remove('active'); t.setAttribute('aria-selected', 'false'); });
      panels.forEach(p => p.classList.remove('active'));
      tab.classList.add('active');
      tab.setAttribute('aria-selected', 'true');
      const target = document.getElementById('panel-' + tab.dataset.tab);
      if (target) target.classList.add('active');
    });
  });
}

/* ============================================================
   8. CLIPBOARD COPY BUTTON
   ============================================================ */
function initClipboard() {
  const btn = document.getElementById('copy-ai-prompt');
  if (!btn) return;
  btn.addEventListener('click', async () => {
    const textEl = document.getElementById('ai-prompt-content');
    if (!textEl) return;
    try {
      await navigator.clipboard.writeText(textEl.textContent);
      const label = btn.querySelector('[data-i18n="install.copy"]') || btn.querySelector('span');
      const original = label.textContent;
      btn.classList.add('copied');
      label.textContent = TRANSLATIONS[currentLang]?.['install.copied'] || 'Copied!';
      setTimeout(() => {
        btn.classList.remove('copied');
        label.textContent = original;
      }, 2000);
    } catch (e) {
      console.warn('Clipboard API unavailable', e);
    }
  });
}

/* ============================================================
   9. STICKY HEADER + NAV HIGHLIGHT
   ============================================================ */
function initHeaderScroll() {
  const header = document.getElementById('site-header');
  if (!header) return;
  window.addEventListener('scroll', () => {
    header.style.borderBottomColor = window.scrollY > 20
      ? 'rgba(212,175,55,0.14)' : 'rgba(255,255,255,0.08)';
  }, { passive: true });
}

function initNavHighlight() {
  const sections = document.querySelectorAll('section[id]');
  const navLinks = document.querySelectorAll('.nav-link[href^="#"]');
  const obs = new IntersectionObserver(entries => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        navLinks.forEach(a => {
          a.style.color = a.getAttribute('href') === `#${entry.target.id}`
            ? 'var(--text-primary)' : '';
        });
      }
    });
  }, { threshold: 0.35 });
  sections.forEach(s => obs.observe(s));
}

/* ============================================================
   10. INIT
   ============================================================ */
document.addEventListener('DOMContentLoaded', () => {
  initI18n();
  initTerminal();
  initScrollReveal();
  initOSDetection();
  initPlatformTabs();
  initInstallTabs();
  initClipboard();
  initHeaderScroll();
  initNavHighlight();
});
