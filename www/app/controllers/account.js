import Controller from '@ember/controller';
import { inject } from '@ember/controller';
import { computed } from '@ember/object';

export default Controller.extend({
  applicationController: inject('application'),
  stats: computed.reads('applicationController'),
  config: computed.reads('applicationController.config'),
  hashrate: computed.reads('applicationController.hashrate'),
  chartOptions: computed("model", {
        get() {
            var now = new Date();
            var e = this,
                t = e.getWithDefault("model.minerCharts", []),
                a = {
                    chart: {
                        backgroundColor: "rgba(255, 255, 255, 0.1)",
                        type: "spline",
                        marginRight: 10,
                        height: 200,
                        events: {
                            load: function() {
                                var self = this;
                                var series = this.series;
                                t = e.getWithDefault("model.minerCharts", []);
                                var a = [];
                                var b = [];

                                // reload chart
                                while (new Date() - series[0].data[series[0].data.length - 1].x > 5*60*1000) {
                                    if (!t || t.length == 0)
                                        break;
                                    t.forEach(function(e) {
                                        var x = new Date(1000 * e.x);
                                        var l = x.toLocaleString();
                                        var y = e.minerLargeHash;
                                        a.push({x: x, y: y, d: l});
                                    });

                                    t.forEach(function(e) {
                                        var x = new Date(1000 * e.x);
                                        var l = x.toLocaleString();
                                        var y = e.minerHash;
                                        b.push({x: x, y: y, d: l});
                                    });
                                    series[0].setData(a, true, {}, true);
                                    series[1].setData(b, true, {}, true);
                                    break;
                                }

                                var chartInterval = setInterval(function() {
                                    var series = self.series;
                                    if (!series) {
                                        clearInterval(chartInterval);
                                        return;
                                    }
                                    var now = new Date();
                                    var shift = false;
                                    if (series[0] && series[0].data.length > 2 && now - series[0].data[0].x > 18*60*60*1000) {
                                        // show 18 hours ~ 15(min) * 74(points) ~ minerChartsNum: 74, minerCharts: "0 */15 ..."
                                        shift = true;
                                    }
                                    // check latest added temporary point and remove tempory added point for less than 5 minutes
                                    if (series[0] && series[0].data.length > 1 &&
                                            series[0].data[series[0].data.length - 1].x - series[0].data[series[0].data.length - 2].x < 5*60*1000) {
                                        series[0].removePoint(series[0].data.length - 1, false, false);
                                        series[1].removePoint(series[1].data.length - 1, false, false);
                                    }
                                    var y = e.getWithDefault("model.hashrate"),
                                        z = e.getWithDefault("model.currentHashrate");
                                    var d = now.toLocaleString();
                                    self.series[0].addPoint({x: now, y: y, d: d}, false, shift);
                                    self.series[1].addPoint({x: now, y: z, d: d}, true, shift);
                                }, e.get('config.highcharts.account.interval') || 120000);
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
                            day: "%m/%d",
                            week: "%m/%d",
                            month: "%b '%y",
                            year: "%Y"
                        }
                    },
                    yAxis: {
                        title: {
                            text: "Hashrate by Account"
                        }
                    },
                    plotLines: [{
                        value: 0,
                        width: 1,
                        color: "#808080"
                    }],
                    legend: {
                        enabled: true,
                        itemStyle: {},
                        item: {
                            text: {
                                style: {}
                            }
                        }
                    },
                    tooltip: {
                        formatter: function() {
                            function scale(v) {
                                var f = v;
                                var units = ['', 'K', 'M', 'G', 'T'];
                                for (var i = 0; i < 5 && f > 1000; i++)  {
                                    f /= 1000;
                                }
                                return f.toFixed(2) + ' ' + units[i];
                            }
                            var h = scale(this.point.y);

                            return "<b>" + this.point.d + "</b><br />" +
                                "<b>Hashrate&nbsp;" + h + "H/s</b>";
                        },

                        useHTML: true
                    },
                    exporting: {
                        enabled: false
                    },
                    plotOptions: {
                        spline: {
                            marker: {
                                enabled: true
                            }
                        }
                    },
                    series: [{
                        color: "#E99002",
                        name: "3 hours average hashrate",
                        data: function() {
                            var a = [];
                            if (null != t) {
                                t.forEach(function(e) {
                                    var x = new Date(1000 * e.x);
                                    var l = x.toLocaleString();
                                    var y = e.minerLargeHash;
                                    a.push({x: x, y: y, d: l});
                                });
                            }
                            var l = now.toLocaleString();
                            var y = e.getWithDefault("model.hashrate");
                            var last = {x: now, y: y, d: l};
                            var interval = e.get('config.highcharts.account.interval') || 120000;
                            if (a.length > 0 && now - a[a.length - 1].x > interval) {
                                a.push(last);
                            }
                            return a;
                        }()
                    }, {
                        name: "30 minutes average hashrate",
                        data: function() {
                            var a = [];
                            if (null != t) {
                                t.forEach(function(e) {
                                    var x = new Date(1000 * e.x);
                                    var l = x.toLocaleString();
                                    var y = e.minerHash;
                                    a.push({x: x, y: y, d: l});
                                });
                            }
                            var l = now.toLocaleString();
                            var y = e.getWithDefault("model.currentHashrate");
                            var last = {x: now, y: y, d: l};
                            var interval = e.get('config.highcharts.account.interval') || 120000;
                            if (a.length > 0 && now - a[a.length - 1].x > interval) {
                                a.push(last);
                            }
                            return a;
                        }()
                    }]
                };
            a.chart.backgroundColor = this.getWithDefault('config.highcharts.account.backgroundColor', "transparent");
            a.title.text = this.getWithDefault('config.highcharts.account.title', '');
            a.yAxis.title.text = this.getWithDefault('config.highcharts.account.ytitle', "Hashrate");
            a.chart.height = this.getWithDefault('config.highcharts.account.height', 200);
            a.xAxis.lineColor = this.getWithDefault('config.highcharts.account.lineColor', "#ccd6eb");
            a.yAxis.lineColor = this.getWithDefault('config.highcharts.account.lineColor', "#ccd6eb");
            a.xAxis.tickColor = this.getWithDefault('config.highcharts.account.tickColor', "#ccd6eb");
            a.yAxis.tickColor = this.getWithDefault('config.highcharts.account.tickColor', "#ccd6eb");
            a.xAxis.gridLineColor = this.getWithDefault('config.highcharts.account.gridLineColor', "#e6e6e6");
            a.yAxis.gridLineColor = this.getWithDefault('config.highcharts.account.gridLineColor', "#e6e6e6");
            a.legend.itemStyle.color = this.getWithDefault('config.highcharts.account.labelColor', "#fff");
            a.legend.item.text.style.color = this.getWithDefault('config.highcharts.account.labelColor', "#fff");
            a.chart.type = this.getWithDefault('config.highcharts.account.type', 'spline');
            var colors = this.getWithDefault('config.highcharts.account.color', ['#e99002', '#1994b8']);
            a.series[0].color = colors[0] || '#e99002';
            a.series[1].color = colors[1] || '#1994b8';
            return a;
        }
    }),
  roundPercent: computed('stats', 'model', {
    get() {
      let percent = this.get('model.roundShares') / this.get('stats.nShares');
      if (!percent) {
        return 0;
      }
      return percent;
    }
  }),

  netHashrate: computed({
    get() {
      return this.get('hashrate');
    }
  }),

  earnPerDay: computed('model', 'stats', {
    get() {
      let reward = this.getWithDefault('stats.blockReward', this.get('config').BlockReward);
      let blocktime = this.getWithDefault('stats.blockTime', this.get('config').BlockTime);
      return 24 * 60 * 60 / blocktime * reward *
        this.getWithDefault('model.hashrate') / this.get('hashrate');
    }
  }),

  earnPerMonth: computed('model', 'stats', {
    get() {
      let reward = this.getWithDefault('stats.blockReward', this.get('config').BlockReward);
      let blocktime = this.getWithDefault('stats.blockTime', this.get('config').BlockTime);
      return 30 * 24 * 60 * 60 / blocktime * reward *
        this.getWithDefault('model.hashrate') / this.get('hashrate');
    }
  })
});
