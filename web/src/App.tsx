import { createSignal, onCleanup, onMount, For } from 'solid-js'
import { createStore } from 'solid-js/store'
import { listBlocks, openEvents } from './lib/api'
import type { BlockListEntry, Snapshot } from './lib/api'
import BlockCard from './components/BlockCard'

export default function App() {
  const [blocks, setBlocks] = createSignal<BlockListEntry[]>([])
  const [snapshots, setSnapshots] = createStore<Record<string, Snapshot>>({})

  onMount(async () => {
    setBlocks(await listBlocks())
    // M1: WS reconnect not handled — page reload restores connection.
    const ws = openEvents((m) => setSnapshots(m.id, m.snapshot))
    onCleanup(() => ws.close())
  })

  return (
    <div class="min-h-screen p-5">
      <h1 class="text-2xl font-semibold mb-4">chop — M1</h1>
      <For each={blocks()}>
        {(b) => <BlockCard block={b} snapshot={snapshots[b.id]} />}
      </For>
    </div>
  )
}
