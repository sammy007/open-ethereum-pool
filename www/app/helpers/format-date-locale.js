import { helper as buildHelper } from '@ember/component/helper';

export function formatDateLocale(ts) {
	let date = new Date(ts * 1000);
  return date.toLocaleString();
}

export default buildHelper(formatDateLocale);
