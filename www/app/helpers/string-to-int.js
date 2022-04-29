import { helper as buildHelper } from '@ember/component/helper';

export function stringToInt(value) {
  return parseInt(value);
}

export default buildHelper(stringToInt);
