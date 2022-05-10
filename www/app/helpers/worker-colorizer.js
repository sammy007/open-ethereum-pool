import Ember from 'ember';

export function workerColorizer(value) {
  let class_name;
  let difference_seconds = (Date.now() / 1000) - value;

  if (difference_seconds >= (60 * 15)) {
    class_name =  "offline-1";
  }

  if (difference_seconds >= (60 * 17)) {
    class_name =  "offline-2";
  }

  if (difference_seconds >= (60 * 20)) {
    class_name =  "offline-3";
  }

  if (difference_seconds >= (60 * 25)) {
    class_name =  "offline-4";
  }

  if (difference_seconds >= (60 * 28)) {
    class_name =  "offline-5";
  }

  return class_name;
}

export default Ember.Helper.helper(workerColorizer);
