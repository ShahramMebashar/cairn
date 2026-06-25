import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: 'cairn',
  description: 'Repo-native task management — one task graph your agents and you share.',
  lang: 'en-US',

  // Project Pages site: https://shahrammebashar.github.io/cairn/
  base: '/cairn/',

  lastUpdated: true,
  cleanUrls: true,

  themeConfig: {
    nav: [
      {
        text: 'Guide',
        link: '/introduction',
        activeMatch: '^/(introduction|installation|quickstart|guides)',
      },
      { text: 'Agents', link: '/agents/', activeMatch: '^/agents/' },
      { text: 'Reference', link: '/reference/mcp-tools', activeMatch: '^/reference/' },
      {
        text: 'v0.x',
        items: [
          { text: 'SPEC (frozen v0 contract)', link: 'https://github.com/ShahramMebashar/cairn/blob/main/SPEC.md' },
          { text: 'Changelog', link: 'https://github.com/ShahramMebashar/cairn/releases' },
        ],
      },
    ],

    sidebar: [
      {
        text: 'Getting started',
        items: [
          { text: 'Introduction', link: '/introduction' },
          { text: 'Installation', link: '/installation' },
          { text: 'Quickstart', link: '/quickstart' },
        ],
      },
      {
        text: 'Core concepts',
        items: [
          { text: 'Task files & config', link: '/guides/task-files' },
          { text: 'The agent loop', link: '/guides/agent-loop' },
          { text: 'Sessions', link: '/guides/sessions' },
          { text: 'Checks & gates', link: '/guides/checks-and-gates' },
        ],
      },
      {
        text: 'Agents',
        items: [
          { text: 'Overview', link: '/agents/' },
          { text: 'Claude Code', link: '/agents/claude' },
          { text: 'Cursor', link: '/agents/cursor' },
          { text: 'Codex', link: '/agents/codex' },
          { text: 'Windsurf', link: '/agents/windsurf' },
          { text: 'OpenCode', link: '/agents/opencode' },
          { text: 'Kilo Code', link: '/agents/kilo' },
          { text: 'Pi', link: '/agents/pi' },
          { text: 'Antigravity', link: '/agents/antigravity' },
        ],
      },
      {
        text: 'Reference',
        items: [
          { text: 'CLI commands', link: '/reference/cli' },
          { text: 'MCP tools', link: '/reference/mcp-tools' },
          { text: 'HTTP API', link: '/reference/http-api' },
          { text: 'Events (SSE)', link: '/reference/events' },
        ],
      },
    ],

    socialLinks: [{ icon: 'github', link: 'https://github.com/ShahramMebashar/cairn' }],

    search: { provider: 'local' },

    editLink: {
      pattern: 'https://github.com/ShahramMebashar/cairn/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'cairn — repo-native task management',
    },
  },
})
