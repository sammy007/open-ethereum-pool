import { helper as buildHelper } from '@ember/component/helper';

export function secondsToMs(value) {
  return value * 1000;
}

export default buildHelper(secondsToMs);
