import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import App from './App'

describe('application shell', () => {
  it('renders every persistent desktop landmark and honest later regions', () => {
    render(<App />)
    expect(screen.getByLabelText('Document bar')).toBeInTheDocument()
    expect(screen.getByLabelText('Component palette')).toBeInTheDocument()
    expect(screen.getByLabelText('Canvas region')).toBeInTheDocument()
    expect(screen.getByLabelText('Blank report page')).toBeInTheDocument()
    expect(screen.getByLabelText('Properties panel')).toBeInTheDocument()
    expect(screen.getByLabelText('Status bar')).toBeInTheDocument()
    expect(screen.getByText('PREVIEW · later')).toBeInTheDocument()
  })

  it('names deferred icon controls and never presents them as working', () => {
    render(<App />)
    const open = screen.getByRole('button', { name: 'Open local template unavailable until local files are added' })
    expect(open).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Save template unavailable until local files are added' })).toBeInTheDocument()
    expect(screen.getByText('Unsaved — local files arrive later')).toBeInTheDocument()
    expect(screen.getByRole('status', { name: 'Offline availability' })).toHaveTextContent('Offline cache unavailable')
  })

  it('announces the checking, ready, and waiting-update lifecycle states', () => {
    const { rerender } = render(<App offlineState="checking" />)
    const status = screen.getByRole('status', { name: 'Offline availability' })
    expect(status).toHaveAttribute('aria-live', 'polite')
    expect(status).toHaveTextContent('Offline cache checking')
    rerender(<App offlineState="ready" />)
    expect(status).toHaveTextContent('Offline ready')
    rerender(<App offlineState="update-available" />)
    expect(status).toHaveTextContent('Update available; current release remains usable')
  })
})
