import { api, type LiveMatch } from './api';

// Polls /api/live for matches in progress. The endpoint reads only from our DB
// (no external API cost); a background sync keeps those scores fresh. Shared by
// the dashboard live panel and the live match page.
const POLL_MS = 60_000;

class LiveStore {
	matches = $state<LiveMatch[]>([]);
	loaded = $state(false);
	private timer: ReturnType<typeof setInterval> | null = null;
	private subscribers = 0;

	private async refresh() {
		try {
			const r = await api.live();
			this.matches = r.matches ?? [];
		} catch {
			/* transient — keep last known list */
		} finally {
			this.loaded = true;
		}
	}

	/** Begin polling. Reference-counted so the dashboard and the live page can
	 *  both subscribe; returns a stop() to call on teardown. */
	start(): () => void {
		this.subscribers += 1;
		if (this.timer === null) {
			void this.refresh();
			this.timer = setInterval(() => void this.refresh(), POLL_MS);
		}
		return () => this.stop();
	}

	private stop() {
		this.subscribers = Math.max(0, this.subscribers - 1);
		if (this.subscribers === 0 && this.timer !== null) {
			clearInterval(this.timer);
			this.timer = null;
		}
	}

	match(id: string): LiveMatch | undefined {
		return this.matches.find((m) => m.id === id);
	}
}

export const liveStore = new LiveStore();
