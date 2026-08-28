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
    expect(screen.getByRole('status')).toHaveTextContent('Unsaved')
  })
})
