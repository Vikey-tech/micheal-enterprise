'use client';
export default function Error({reset}){return <div className="shell py-20"><h2>Something went wrong.</h2><p className="muted my-6">Please try again. If the issue continues, contact the dealership.</p><button className="btn" onClick={reset}>Try again</button></div>}
