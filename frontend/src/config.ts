export const siteName = 'bAI'
export const siteTagline = 'The Autonomous Terminal AI Coding Agent'
export const siteTitle = `${siteName} — ${siteTagline}`
export const repoUrl = 'https://github.com/biisal/bai'
export const providers = [
  { name: 'OpenAI', model: 'gpt-6-astra' },
  { name: 'Anthropic', model: 'claude-fable-5-1' },
  { name: 'DeepSeek', model: 'deepseek-v4-pro' },
  { name: 'Qwen', model: 'qwen-max' },
  { name: 'Groq', model: 'llama-3.3-70b' },
  { name: 'OpenCode', model: 'grok-code' },
]
export const installUrl = 'https://raw.githubusercontent.com/biisal/bai/main/install'
export const installCmd = `curl -fsSL ${installUrl} | bash`
