import Ember from 'ember';

export default Ember.Controller.extend({
  applicationController: Ember.inject.controller('application'),
  config: Ember.computed.reads('applicationController.config'),
  netstats: Ember.computed.reads('applicationController'),
  stats: Ember.computed.reads('applicationController.model.stats'),

  chartOptions: Ember.computed("model", {
        get() {
            var now = new Date();
            var e = this,
                t = e.getWithDefault("model.minerCharts"),
                a = {
                    chart: {
                        backgroundColor: "rgba(255, 255, 255, 0.1)",
                        type: "spline",
                        marginRight: 10,
                        height: 200,
                        events: {
                            load: function() {
                                var self = this;
                                setInterval(function() {
                                    var series = self.series;
                                    if (!series) {
                                        return; // FIXME
                                    }
                                    var now = new Date();
                                    var shift = false;
                                    if (series && series[0] && series[0].data && series[0].data[0] && now - series[0].data[0].x > 6*60*60*1000) {
                                        shift = true;
                                    }
                                    var y = e.getWithDefault("model.hashrate"),
                                        z = e.getWithDefault("model.currentHashrate");
                                    var d = now.toLocaleString();
                                    self.series[0].addPoint({x: now, y: y, d: d}, true, shift);
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
                            day: "%e. %b",
                            week: "%e. %b",
                            month: "%b '%y",
                            year: "%Y"
                        }
                    },
                    yAxis: {
                        title: {
                            text: "Hashrate by Account"
                        },
                        //softMin: e.getWithDefault("model.currentHashrate") / 1000000,
                        //softMax: e.getWithDefault("model.currentHashrate") / 1000000,
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
            return a;
        }
    })
});
