import { Show } from 'solid-js'
import type { BlockListEntry, LinkStats, Snapshot } from '../lib/api'
import LinkSlider from './LinkSlider'

const blockTypeLink = 'link'

export default function BlockCard(props: { block: BlockListEntry; snapshot?: Snapshot }) {
  const linkStats = (): LinkStats | undefined =>
    props.block.type === blockTypeLink ? (props.snapshot?.stats as LinkStats | undefined) : undefined

  return (
    <div class="bg-card rounded-md p-4 mb-3">
      <div>
        <strong>{props.block.id}</strong>{' '}
        <span class="text-muted">({props.block.type})</span>
      </div>
      <pre class="text-muted text-xs my-2">
        {JSON.stringify(props.snapshot?.stats ?? {}, null, 2)}
      </pre>
      <Show when={linkStats()}>
        {(stats) => <LinkSlider id={props.block.id} initialLoss={stats().controls.loss} />}
      </Show>
    </div>
  )
}
