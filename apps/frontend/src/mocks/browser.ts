import { setupWorker } from 'msw/browser'
import { browserHandlers } from './handlers/index'

export const worker = setupWorker(...browserHandlers)
