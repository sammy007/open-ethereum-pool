import Ember from 'ember';

export default Ember.Controller.extend({
  applicationController: Ember.inject.controller('application'),
  config: Ember.computed.reads('applicationController.config'),
  stats: Ember.computed.reads('applicationController.model.stats'),
  hashrate: Ember.computed.reads('applicationController.hashrate'),
  chartOptions: Ember.computed("model.hashrate", {
        get() {
            var e = this,
                t = e.getWithDefault("model.minerCharts"),
                a = {
                    chart: {
                        backgroundColor: "rgba(255, 255, 255, 0.1)",
                        type: "spline",
                        marginRight: 10,
                        height: 400,
                        events: {
                            load: function() {
                                var series = this.series[0];
                                setInterval(function() {
                                    var now = new Date();
                                    var shift = false;
                                    if (now - series.data[0].x > 6*60*60*1000) {
                                        shift = true;
                                    }

                                    var x = now,
                                        y = e.getWithDefault("model.currentHashrate");
                                    var d = x.toLocaleString();
                                    series.addPoint({x:x, y:y, d:d}, true, shift);
                                }, 10000);
                            }
                        }
                    },
                    title: {
                        text: ""
                    },
                    xAxis: {
                        ordinal: false,
                        type: "datetime",
                        dateTimeLabelFormats: {
                            millisecond: "%H:%M:%S",
                            second: "%H:%M:%S",
                            minute: "%H:%M",
                            hour: "%H:%M",
                            day: "%e. %b",
                            week: "%e. %b",
                            month: "%b '%y",
                            year: "%Y"
                        }
                    },
                    yAxis: {
                        title: {
                            text: "HashRate"
                        },
                        min: 0
                    },
                    plotLines: [{
                        value: 0,
                        width: 1,
                        color: "#808080"
                    }],
                    legend: {
                        enabled: true
                    },
                    tooltip: {
                        formatter: function() {
                            return this.y > 1000000000000 ? "<b>" + this.point.d + "<b><br>Hashrate&nbsp;" + (this.y / 1000000000000).toFixed(2) + "&nbsp;TH/s</b>" : this.y > 1000000000 ? "<b>" + this.point.d + "<b><br>Hashrate&nbsp;" + (this.y / 1000000000).toFixed(2) + "&nbsp;GH/s</b>" : this.y > 1000000 ? "<b>" + this.point.d + "<b><br>Hashrate&nbsp;" + (this.y / 1000000).toFixed(2) + "&nbsp;MH/s</b>" : "<b>" + this.point.d + "<b><br>Hashrate&nbsp;<b>" + this.y.toFixed(2) + "&nbsp;H/s</b>";

                        },

                        useHTML: true
                    },
                    exporting: {
                        enabled: false
                    },
                    series: [{
                        color: "#E99002",
                        name: "Average hashrate",
                        data: function() {
                            var e, a = [];
                            if (null != t) {
                                for (e = 0; e <= t.length - 1; e += 1) {
                                    var n = 0,
                                        r = 0,
                                        l = 0;
                                    r = new Date(1e3 * t[e].x);
                                    l = r.toLocaleString();
                                    n = t[e].minerLargeHash;
                                    a.push({
                                        x: r,
                                        d: l,
                                        y: n
                                    });
                                }
                            } else {
                                a.push({
                                x: 0,
                                d: 0,
                                y: 0
                            });
                            }
                            return a;
                        }()
                    }, {
                        name: "Current hashrate",
                        data: function() {
                            var e, a = [];
                            if (null != t) {
                                for (e = 0; e <= t.length - 1; e += 1) {
                                    var n = 0,
                                        r = 0,
                                        l = 0;
                                    r = new Date(1e3 * t[e].x);
                                    l = r.toLocaleString();
                                    n = t[e].minerHash;
                                    a.push({
                                        x: r,
                                        d: l,
                                        y: n
                                    });
                                }
                            } else {
                                a.push({
                                x: 0,
                                d: 0,
                                y: 0
                                });
                            }
                            return a;
                        }()
                    }]
                };
            a.title.text = this.get('config.highcharts.account.title') || "";
            a.yAxis.title.text = this.get('config.highcharts.account.ytitle') || "Hashrate";
            a.chart.height = this.get('config.highcharts.account.height') || 300;
            a.chart.type = this.get('config.highcharts.account.type') || 'spline';
            var colors = this.get('config.highcharts.account.color');
            a.series[0].color = colors[0] || '#e99002';
            a.series[1].color = colors[1] || '#1994b8';
            return a;
        }
    }),
  roundPercent: Ember.computed('stats', 'model', {
    get() {
      var percent = this.get('model.roundShares') / this.get('stats.roundShares');
      if (!percent) {
        return 0;
      }
      return percent;
    }
  }),

  netHashrate: Ember.computed({
    get() {
      return this.get('hashrate');
    }
  }),

  earnPerDay: Ember.computed('model', {
    get() {
      return 24 * 60 * 60 / this.get('config').BlockTime * this.get('config').BlockReward *
        this.getWithDefault('model.hashrate') / this.get('hashrate');
    }
  })
});
