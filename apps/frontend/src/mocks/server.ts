import { setupServer } from 'msw/node'
import { serverHandlers } from './handlers/index'

export const server = setupServer(...serverHandlers)
