import Ember from 'ember';

export default Ember.Controller.extend({
  applicationController: Ember.inject.controller('application'),
  stats: Ember.computed.reads('applicationController'),
  config: Ember.computed.reads('applicationController.config'),
  settings: Ember.computed.reads('applicationController.model.settings'),

  // try to read some settings from the model.settings
  PayoutThreshold: Ember.computed('settings', {
    get() {
      var threshold = this.get('settings.PayoutThreshold');
      if (threshold) {
        // in shannon (10**9)
        return threshold / 1000000000;
      }
      return this.get('config').PayoutThreshold;
    }
  }),

  PayoutInterval: Ember.computed('settings', {
    get() {
      var interval = this.get('settings.PayoutInterval');
      if (interval) {
        return interval;
      }
      return this.get('config').PayoutInterval;
    }
  }),

  PoolFee: Ember.computed('settings', {
    get() {
      var poolfee = this.get('settings.PoolFee');
      if (poolfee) {
        return poolfee + '%';
      }
      return this.get('config').PoolFee;
    }
  }),

	cachedLogin: Ember.computed('login', {
    get() {
      return this.get('login') || Ember.$.cookie('login');
    },
    set(key, value) {
      Ember.$.cookie('login', value);
      this.set('model.login', value);
      return value;
    }
  }),
  chartOptions: Ember.computed("model.hashrate", {
        get() {
            var e = this,
                t = e.getWithDefault("stats.model.poolCharts"),
                a = {
                    chart: {
                        backgroundColor: "rgba(255, 255, 255, 0.1)",
                        type: "spline",
                        height: 300,
                        marginRight: 10,
                        events: {
                            load: function() {
                                var series = this.series[0];
                                setInterval(function() {var x = (new Date).getTime(), y = e.getWithDefault("model.Hashrate") / 1000000; series.addPoint([x, y], true, true)}, 1090000000)
                            }
                        }
                    },
                    title: {
                        text: "Our pool's hashrate"
                    },
                    xAxis: {
                        labels: {
                            style: {
                                color: "#000"
                            }
                        },
                        ordinal: false,
                        type: "datetime"
                    },
                    yAxis: {
                        title: {
                            text: "HASHRATE",
                            style: {
                                color: "#000"
                            }
                        },
                        min: 0,
                        labels: {
                            style: {
                                color: "#000"
                            }
                        }
                    },
                    plotLines: [{
                        value: 0,
                        width: 1,
                        color: "#000"
                    }],
                    legend: {
                        enabled: false
                    },
                    tooltip: {
                        formatter: function() {
                            return this.y > 1000000000000 ? "<b>" + this.point.d + "<b><br>Hashrate&nbsp;" + (this.y / 1000000000000).toFixed(2) + "&nbsp;TH/s</b>" : this.y > 1000000000 ? "<b>" + this.point.d + "<b><br>Hashrate&nbsp;" + (this.y / 1000000000).toFixed(2) + "&nbsp;GH/s</b>" : this.y > 1000000 ? "<b>" + this.point.d + "<b><br>Hashrate&nbsp;" + (this.y / 1000000).toFixed(2) + "&nbsp;MH/s</b>" : "<b>" + this.point.d + "<b><br>Hashrate<b>&nbsp;" + this.y.toFixed(2) + "&nbsp;H/s</b>"
                        },
                        useHTML: true
                    },
                    exporting: {
                        enabled: false
                    },
                    series: [{
                        color: "#15BD27",
                        name: "Hashrate",
                        data: function() {
                            var e, a = [];
                            if (null != t)
                                for (e = 0; e <= t.length - 1; e += 1) {
                                    var n = 0,
                                        r = 0,
                                        l = 0;
                                    r = new Date(1e3 * t[e].x), l = r.toLocaleString(), n = t[e].y, a.push({
                                        x: r,
                                        d: l,
                                        y: n
                                    })
                                } else a.push({
                                x: 0,
                                d: 0,
                                y: 0
                            });
                            return a
                        }()
                    }]
                };
            return a
        }
    })
});
