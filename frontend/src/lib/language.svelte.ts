import { browser } from '$app/environment';

export type LanguageCode = 'nn' | 'en';

const STORAGE_KEY = 'language';

export function isLanguageCode(value: unknown): value is LanguageCode {
	return value === 'nn' || value === 'en';
}

function readStoredLanguage(): LanguageCode {
	if (!browser) return 'en';
	const stored = localStorage.getItem(STORAGE_KEY);
	return isLanguageCode(stored) ? stored : 'en';
}

class LanguageStore {
	code = $state<LanguageCode>(readStoredLanguage());

	get resolved() {
		return this.code;
	}

	get locale() {
		// `nn-NO` falls back inconsistently in some browsers; `no-NO`
		// keeps Norwegian date/time formatting stable while UI copy stays Nynorsk.
		return this.code === 'nn' ? 'no-NO' : 'en-US';
	}

	get isEnglish() {
		return this.code === 'en';
	}

	set(next: LanguageCode) {
		this.code = next;
		if (browser) localStorage.setItem(STORAGE_KEY, next);
	}

	toggle() {
		this.set(this.code === 'nn' ? 'en' : 'nn');
	}
}

export const language = new LanguageStore();