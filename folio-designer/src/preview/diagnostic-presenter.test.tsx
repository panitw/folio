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

  it('announces a failed local render and always returns, including when location data is absent', () => {
    const returned = vi.fn()
    render(<PreviewFailure error={{ code: 'RENDER_INVALID', message: 'The template could not be processed', dataPath: 'items[0]' }} onReturn={returned} />)
    expect(screen.getByLabelText('Local render failure')).toHaveTextContent('Local PDF render failed')
    expect(screen.getByLabelText('Local render failure')).toHaveAttribute('aria-atomic', 'true')
    fireEvent.click(screen.getByRole('button', { name: 'Return to Design' }))
    expect(returned).toHaveBeenCalledOnce()
  })
})
