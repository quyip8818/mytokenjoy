import { Component, type ReactNode } from 'react'

interface State {
  hasError: boolean
}

export class RouteErrorBoundary extends Component<{ children: ReactNode }, State> {
  state: State = { hasError: false }

  static getDerivedStateFromError(): State {
    return { hasError: true }
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex h-full flex-col items-center justify-center gap-4">
          <p className="text-lg text-muted-foreground">页面出错了</p>
          <button
            onClick={() => this.setState({ hasError: false })}
            className="rounded-md bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90"
          >
            重试
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
