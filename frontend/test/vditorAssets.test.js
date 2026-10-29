import test from 'node:test'
import assert from 'node:assert/strict'
import { vditorAssets } from '../vite/vditorAssets.js'

test('production bundles contain Markdown runtime, editor resources and preview themes', async () => {
  const files = new Map()
  await vditorAssets().generateBundle.call({
    emitFile(asset) {
      files.set(asset.fileName, asset.source)
    }
  })
  for (const path of [
    'js/lute/lute.min.js',
    'js/i18n/zh_CN.js',
    'js/icons/ant.js',
    'js/highlight.js/highlight.min.js',
    'js/mermaid/mermaid.min.js',
    'css/content-theme/light.css'
  ]) {
    const content = files.get(`vendor/vditor/dist/${path}`)
    assert.ok(content?.length > 0, `Missing locally served rendering resource: ${path}`)
  }
})
