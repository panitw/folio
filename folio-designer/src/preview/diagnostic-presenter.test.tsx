import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { PreviewDiagnostics, PreviewFailure } from './diagnostic-presenter'

const warning = { severity: 'warning' as const, code: 'CONTENT_CLIPPED', elementId: 'e7', dataPath: 'bands.content.e7', message: 'Content was clipped' }

describe('diagnostic presenter', () => {
  it('renders producer facts with shape, mono location, dismiss, and locate controls', () => {
    const dismiss = vi.fn(); const locate = vi.fn()
    render(<PreviewDiagnostics diagnostics={[warning]} dismissed={new Set()} onDismiss={dismiss} onLocate={locate} />)
    expect(screen.getByLabelText('Render diagnostics')).toBeInTheDocument()
    expect(screen.getByRole('status')).toHaveTextContent('1 render warning available: CONTENT_CLIPPED.')
    expect(screen.getByText('warning · CONTENT_CLIPPED')).toBeInTheDocument()
    expect(screen.getByText('e7 · bands.content.e7')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Locate in Design' }))
    expect(locate).toHaveBeenCalledWith(warning)
    fireEvent.click(screen.getByRole('button', { name: 'Dismiss CONTENT_CLIPPED diagnostic' }))
    expect(dismiss).toHaveBeenCalledOnce()
  })

  it('keeps the one admitted-generation announcement stable while local cards are dismissed', () => {
    const { rerender } = render(<PreviewDiagnostics diagnostics={[warning, { ...warning, code: 'ROW_TOO_TALL' }]} dismissed={new Set()} onDismiss={vi.fn()} onLocate={vi.fn()} />)
    const status = screen.getByRole('status')
    expect(status).toHaveTextContent('2 render warnings available: CONTENT_CLIPPED, ROW_TOO_TALL.')
    rerender(<PreviewDiagnostics diagnostics={[warning, { ...warning, code: 'ROW_TOO_TALL' }]} dismissed={new Set(['0:warning:CONTENT_CLIPPED:e7:bands.content.e7:Content was clipped'])} onDismiss={vi.fn()} onLocate={vi.fn()} />)
    expect(status).toHaveTextContent('2 render warnings available: CONTENT_CLIPPED, ROW_TOO_TALL.')
  })

  it('does not invent a locate target when the engine supplied no element id', () => {
    render(<PreviewDiagnostics diagnostics={[{ ...warning, elementId: '', dataPath: 'page.margin.top' }]} dismissed={new Set()} onDismiss={vi.fn()} onLocate={vi.fn()} />)
    expect(screen.queryByRole('button', { name: 'Locate in Design' })).not.toBeInTheDocument()
    expect(screen.getByText('page.margin.top')).toHaveClass('diagnostic-location')
  })

  it('announces opaque producer facts independently with an error shape and named keyboard controls', () => {
    const returned = vi.fn(); const retried = vi.fn()
    render(<PreviewFailure error={{ code: 'RENDER_INVALID', message: 'The template could not be processed', elementId: 'e7', dataPath: 'items[0]' }} onRetry={retried} onReturn={returned} />)
    const failure = screen.getByLabelText('Local render failure')
    expect(failure).toHaveTextContent('Local PDF render failed')
    expect(failure).toHaveTextContent('Render failure · RENDER_INVALID')
    expect(failure).toHaveTextContent('Element ID: e7')
    expect(failure).toHaveTextContent('Data path: items[0]')
    expect(failure).toHaveAttribute('role', 'alert')
    expect(failure).toHaveAttribute('aria-atomic', 'true')
    expect(failure.querySelector('.preview-failure-marker')).toHaveTextContent('■')
    const retry = screen.getByRole('button', { name: 'Retry preview' })
    retry.focus()
    expect(document.activeElement).toBe(retry)
    fireEvent.click(retry)
    expect(retried).toHaveBeenCalledOnce()
    fireEvent.click(screen.getByRole('button', { name: 'Return to Design' }))
    expect(returned).toHaveBeenCalledOnce()
  })

  it('keeps path-only and location-free failures readable without manufacturing an id', () => {
    const { rerender } = render(<PreviewFailure error={{ code: 'RENDER_INVALID', message: 'The template could not be processed', dataPath: 'items[0]' }} onRetry={vi.fn()} onReturn={vi.fn()} />)
    expect(screen.getByText('Data path: items[0]')).toHaveClass('diagnostic-location')
    expect(screen.queryByText(/Element ID:/)).not.toBeInTheDocument()
    rerender(<PreviewFailure error={{ code: 'RENDER_INVALID', message: 'The template could not be processed' }} onRetry={vi.fn()} onReturn={vi.fn()} />)
    expect(screen.queryByText(/Element ID:|Data path:/)).not.toBeInTheDocument()
  })
})
