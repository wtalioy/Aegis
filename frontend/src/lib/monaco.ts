// Monaco Editor Configuration - Phase 4
import * as monaco from 'monaco-editor'

export function configureMonaco() {
  // Configure YAML language support
  monaco.languages.register({ id: 'yaml' })
  
  monaco.languages.setMonarchTokensProvider('yaml', {
    tokenizer: {
      root: [
        [/---/, 'delimiter'],
        [/\.\.\./, 'delimiter'],
        [/(\w+)(\s*)(:)/, ['key', 'white', 'delimiter']],
        [/("|')(.*?)(\1)/, 'string'],
        [/\d+/, 'number'],
        [/#.*$/, 'comment']
      ]
    }
  })

  // Set theme
  monaco.editor.defineTheme('aegis-dark', {
    base: 'vs-dark',
    inherit: true,
    rules: [
      { token: 'key', foreground: '569cd6' },
      { token: 'string', foreground: 'ce9178' },
      { token: 'number', foreground: 'b5cea8' },
      { token: 'comment', foreground: '6a9955', fontStyle: 'italic' }
    ],
    colors: {
      'editor.background': '#1a1a1a',
      'editor.foreground': '#d4d4d4'
    }
  })

  monaco.editor.setTheme('aegis-dark')
}

export function createYamlEditor(container: HTMLElement, value: string): monaco.editor.IStandaloneCodeEditor {
  return monaco.editor.create(container, {
    value,
    language: 'yaml',
    theme: 'aegis-dark',
    automaticLayout: true,
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    fontSize: 13,
    lineNumbers: 'on',
    roundedSelection: false,
    cursorStyle: 'line',
    readOnly: false,
    wordWrap: 'on'
  })
}

