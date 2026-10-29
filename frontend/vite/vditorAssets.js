import { createReadStream } from 'node:fs'
import { readdir, readFile, stat } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, extname, join, relative, resolve, sep } from 'node:path'

const require = createRequire(import.meta.url)
const packageRoot = dirname(require.resolve('vditor/package.json'))
const assetRoot = join(packageRoot, 'dist')
const contentTypes = {
  '.js': 'text/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.gif': 'image/gif',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
  '.wasm': 'application/wasm'
}

async function* filesIn(directory) {
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name)
    if (entry.isDirectory()) yield* filesIn(path)
    else if (entry.isFile()) yield path
  }
}

// 开发时直接提供依赖包中的资源，构建时将同一版本的运行资源写入产物。
export function vditorAssets() {
  let base = '/'
  return {
    name: 'mixm-vditor-assets',
    configResolved(config) {
      base = config.base
    },
    configureServer(server) {
      server.middlewares.use(`${base}vendor/vditor/dist`, async (req, res, next) => {
        try {
          const pathname = decodeURIComponent(new URL(req.url, 'http://localhost').pathname)
          const file = resolve(assetRoot, `.${pathname}`)
          if (!file.startsWith(assetRoot + sep)) {
            res.statusCode = 404
            res.end()
            return
          }
          const info = await stat(file)
          if (!info.isFile()) return next()
          res.setHeader('Content-Type', contentTypes[extname(file)] || 'application/octet-stream')
          res.setHeader('Content-Length', info.size)
          if (req.method === 'HEAD') res.end()
          else createReadStream(file).on('error', next).pipe(res)
        } catch (error) {
          if (error.code === 'ENOENT') return next()
          next(error)
        }
      })
    },
    async generateBundle() {
      for (const directory of ['js', 'css', 'images']) {
        for await (const file of filesIn(join(assetRoot, directory))) {
          this.emitFile({
            type: 'asset',
            fileName: `vendor/vditor/dist/${relative(assetRoot, file).split(sep).join('/')}`,
            source: await readFile(file)
          })
        }
      }
    }
  }
}
