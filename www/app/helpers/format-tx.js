import { helper as buildHelper } from '@ember/component/helper';

export function formatTx(value) {
  return value[0].substring(2, 26) + "..." + value[0].substring(42);
}

export default buildHelper(formatTx);
