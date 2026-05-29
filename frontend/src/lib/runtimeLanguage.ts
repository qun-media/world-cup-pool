export type RuntimeLanguageCode = 'nn' | 'en';

const STORAGE_KEY = 'language';

export function readRuntimeLanguage(): RuntimeLanguageCode {
	if (typeof window !== 'undefined') {
		try {
			return localStorage.getItem(STORAGE_KEY) === 'en' ? 'en' : 'nn';
		} catch {
			const htmlLang = document.documentElement.lang.toLowerCase();
			return htmlLang.startsWith('en') ? 'en' : 'nn';
		}
	}
	return 'nn';
}

export function isRuntimeEnglish() {
	return readRuntimeLanguage() === 'en';
}

export function readRuntimeLocale() {
	return isRuntimeEnglish() ? 'en-US' : 'nn-NO';
}