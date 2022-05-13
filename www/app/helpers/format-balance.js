import { helper as buildHelper } from '@ember/component/helper';

export function formatBalance(value) {
  value = value * 0.000000001;
  return value.toFixed(8);
}

export default buildHelper(formatBalance);
