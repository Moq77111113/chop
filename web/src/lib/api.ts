export type BlockListEntry = { id: string; type: string }
export type Snapshot = { status: string; stats: unknown; ts: number }
export type SnapshotMsg = { type: 'snapshot'; id: string; snapshot: Snapshot }

export type LinkControls = {
  loss: number
  latency_ms: number
  jitter_ms: number
  bandwidth_kbps: number
}

export type LinkStats = {
  packets_in: number
  packets_out: number
  packets_dropped: number
  up_since_ms: number
  controls: LinkControls
}

export const apiPaths = {
  blocks: '/api/blocks',
  controls: (id: string) => `/api/blocks/${id}/controls`,
  events: '/api/events',
} as const

export async function listBlocks(): Promise<BlockListEntry[]> {
  const r = await fetch(apiPaths.blocks)
  if (!r.ok) throw new Error(`listBlocks: ${r.status}`)
  return r.json()
}

export async function patchControls(id: string, controls: object): Promise<void> {
  const r = await fetch(apiPaths.controls(id), {
    method: 'PATCH',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(controls),
  })
  if (!r.ok) throw new Error(`patchControls: ${r.status}`)
}

export function openEvents(onMsg: (m: SnapshotMsg) => void): WebSocket {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const ws = new WebSocket(`${proto}://${location.host}${apiPaths.events}`)
  ws.onmessage = (e) => {
    try {
      onMsg(JSON.parse(e.data))
    } catch {
      // M1: malformed payload silently dropped; supervisor controls schema.
    }
  }
  return ws
}
