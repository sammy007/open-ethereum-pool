import { helper as buildHelper } from '@ember/component/helper';

export function formatHashrate(params/*, hash*/) {
  let hashrate = params[0];
  let i = 0;
  let units = ['H', 'KH', 'MH', 'GH', 'TH', 'PH'];
  while (hashrate > 1000) {
    hashrate = hashrate / 1000;
    i++;
  }
  return hashrate.toFixed(2) + ' ' + units[i];
}

export default buildHelper(formatHashrate);
