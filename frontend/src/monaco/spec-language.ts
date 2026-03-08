import * as monaco from 'monaco-editor'

export function registerSpecLanguage() {
  monaco.languages.register({ id: 'rpmspec' })

  monaco.languages.setMonarchTokensProvider('rpmspec', {
    keywords: [
      'Name',
      'Version',
      'Release',
      'Summary',
      'License',
      'URL',
      'Source0',
      'Patch0',
      'BuildArch',
      'BuildRoot',
      'BuildRequires',
      'Requires',
      'Provides',
      'Obsoletes',
      'Conflicts',
      '%description',
      '%prep',
      '%build',
      '%install',
      '%clean',
      '%files',
      '%changelog',
      '%package',
      '%post',
      '%postun',
      '%pre',
      '%preun',
      '%setup',
      '%patch',
      '%define',
      '%global',
      '%include',
      '%if',
      '%else',
      '%endif',
      '%ifarch',
      '%ifos',
      '%dist',
      '%autorelease',
      '%autochangelog',
    ],

    operators: ['>=', '<=', '>', '<', '=', '!='],

    symbols: /[><=!]+/,

    tokenizer: {
      root: [
        [/#.*/, 'comment'],
        [/%\w+/, 'keyword'],
        [/\$\w+/, 'variable'],
        [/\$\{\w+\}/, 'variable'],
        [/%\{\w+\}/, 'variable'],
        [/@\w+/, 'variable'],
        [/[><=!]+/, 'operator'],
        [/\d+\.\d+\.\d+/, 'number'],
        [/\d+/, 'number'],
        [/"([^"]*)"/, 'string'],
        [/'([^']*)'/, 'string'],
        [/[a-zA-Z_]\w*/, {
          cases: {
            '@keywords': 'keyword',
            '@default': 'identifier',
          },
        }],
      ],
    },
  })

  monaco.languages.setLanguageConfiguration('rpmspec', {
    comments: {
      lineComment: '#',
    },
    brackets: [
      ['{', '}'],
      ['[', ']'],
      ['(', ')'],
    ],
    autoClosingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
      { open: "'", close: "'" },
    ],
    surroundingPairs: [
      { open: '{', close: '}' },
      { open: '[', close: ']' },
      { open: '(', close: ')' },
      { open: '"', close: '"' },
      { open: "'", close: "'" },
    ],
  })
}
