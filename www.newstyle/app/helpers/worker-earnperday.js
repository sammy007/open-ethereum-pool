import Ember from 'ember';
import config from '../config/environment';


export function workerEarnperday(hashrates) {
  return 24 * 60 * 60 / config.APP.BlockTime * (hashrates[0] / hashrates[1]) * config.APP.BlockReward;
}

export default Ember.Helper.helper(workerEarnperday);
