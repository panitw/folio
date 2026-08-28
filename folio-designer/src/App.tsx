import './App.css'

const paletteItems = ['Text', 'Image', 'Table', 'Line', 'Rectangle']

function Icon({ name }: { name: 'open' | 'save' }) {
  return <svg aria-hidden="true" className="icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.25"><path d={name === 'open' ? 'M2 5.5h4l1.2-2h6.8v9H2z M2 5.5h12' : 'M3 2h8l2 2v10H3z M5 2v4h6V2 M5 12h6'} /></svg>
}

function DeferredIconButton({ label, icon }: { label: string; icon: 'open' | 'save' }) {
  return <button className="icon-button" type="button" disabled aria-describedby="future-features"><Icon name={icon} /><span className="sr-only">{label}</span></button>
}

export default function App() {
  return <div className="app-shell" aria-label="Folio designer application shell">
    <header className="document-bar" aria-label="Document bar">
      <span className="brand">FOLIO</span><span className="document-name">Untitled template</span><span className="status-dot" aria-hidden="true" /><span className="status-copy" role="status">Unsaved — local files arrive later</span>
      <div className="document-actions" aria-label="File actions unavailable until local files are added"><DeferredIconButton label="Open local template unavailable until local files are added" icon="open" /><DeferredIconButton label="Save template unavailable until local files are added" icon="save" /></div>
      <span className="later-control" aria-label="Page setup is coming later">A4 · portrait</span>
      <div className="mode-switch" aria-label="Designer mode"><span className="mode-active">DESIGN</span><span className="mode-disabled">PREVIEW · later</span></div>
    </header>
    <div className="workbench" id="future-features">
      <nav className="palette-rail" aria-label="Component palette"><p className="section-label">PALETTE</p>{paletteItems.map((item) => <div className="palette-item" key={item}><span className="palette-icon" aria-hidden="true" />{item}<kbd>later</kbd></div>)}<p className="honest-note">Placement arrives later.</p></nav>
      <main className="canvas-region" aria-label="Canvas region"><section className="page-surface" aria-label="Blank report page"><p className="page-eyebrow">UNTITLED TEMPLATE</p><h1>Report page</h1><p className="page-copy">Canvas editing and bands arrive in later stories.</p><span className="page-placeholder">A4 · 210 × 297 mm</span></section></main>
      <aside className="properties-panel" aria-label="Properties panel"><p className="section-label">PROPERTIES</p><p className="panel-heading">No selection</p><div className="property-field"><span>POSITION</span><code>—</code></div><div className="property-field"><span>BINDING</span><code>unavailable</code></div><p className="honest-note">Select a component after canvas editing is available.</p></aside>
    </div>
    <footer className="status-bar" aria-label="Status bar"><span>LOCAL SHELL</span><code>NO TEMPLATE LOADED</code><span className="status-spacer" /><code>DESIGN MODE</code></footer>
  </div>
}
