import { helper as buildHelper } from '@ember/component/helper';

export function formatBalance(params) {
  let [value, fixed = 8, figures = 9] = params;
  if (!value) {
    return "0.0";
  }
  value = value * 0.000000001;
  return Number(value.toFixed(fixed)).toPrecision(figures).replace(/0+$/, '').replace(/\.$/, '.0');
}

export default buildHelper(formatBalance);
