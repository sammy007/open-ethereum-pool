import { helper as buildHelper } from '@ember/component/helper';

export function withMetricPrefix(params/*, hash*/) {
  let n = params[0];

  if (n < 1000) {
    return n;
  }

  let i = 0;
  let units = ['K', 'M', 'G', 'T', 'P'];
  while (n > 1000) {
    n = n / 1000;
    i++;
  }
  return n.toFixed(3) + ' ' + units[i - 1];
}

export default buildHelper(withMetricPrefix);
