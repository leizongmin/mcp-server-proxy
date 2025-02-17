import { handle } from 'hono/service-worker'
import app from './app.js'

self.addEventListener('fetch', handle(app));
