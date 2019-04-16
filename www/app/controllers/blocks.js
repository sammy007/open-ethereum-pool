import Controller from '@ember/controller';
import { inject } from '@ember/controller';
import { computed } from '@ember/object';

export default Controller.extend({
  applicationController: inject('application'),
  stats: computed.reads('applicationController'),
  config: computed.reads('applicationController.config'),
  settings: computed.reads('applicationController.model.settings'),
  lastData: null,

  BlockUnlockDepth: computed('settings', {
    get() {
      var depth = this.get('settings.BlockUnlockDepth');
      if (depth) {
        return depth;
      }
      return this.get('config').BlockUnlockDepth;
    }
  }),

  chartOptions: computed("model.luckCharts", 'stats', {
        get() {
            var e = this,
                t = e.getWithDefault("model.luckCharts", []),
                a = {
                    colors: ['#f45b5b', '#8085e9', '#8d4654', '#7798BF', '#aaeeee',
                            '#ff0066', '#eeaaee', '#55BF3B', '#DF5353', '#7798BF', '#aaeeee'],
                    chart: {
                        backgroundColor: "rgba(255, 255, 255, 0.1)",
                        marginRight: 100,
                        height: 200,
                        events: {
                            load: function() {
                                var series = this.series;

                                var chartInterval = setInterval(function() {
                                    if (!series || !series[0]) {
                                        clearInterval(chartInterval);
                                        return;
                                    }
                                    var now = new Date();
                                    t = e.getWithDefault("model.luckCharts", []);
                                    // save lastData or reload LastData
                                    var lastData = e.get('lastData');
                                    if (!lastData) {
                                        lastData = series[0].data[series[0].data.length - 1];
                                        e.set('lastData', lastData);
                                    }
                                    var dataLast = t && t.length ? t[t.length - 1] : null;

                                    var check = dataLast ? dataLast.x - parseInt(lastData.x / 1000) : 0;
                                    check = check < 0 ? -check : check;
                                    if (check < 5) {
                                        // check datapoint shift
                                        var shift = false;
                                        if (series[0].data.length > 2 && now - series[0].data[0].x > 15*24*60*60*1000) {
                                            // show 15 days
                                            shift = true;
                                        }

                                        // partially update chart
                                        if (series[0].data.length > 1 && series[0].data[series[0].data.length - 1].x - series[0].data[series[0].data.length - 2].x < 5*60*1000) {
                                            // remove temporary added point
                                            series[0].removePoint(series[0].data.length - 1, false, false);
                                            series[1].removePoint(series[1].data.length - 1, false, false);
                                        }

                                        // simply update the last point
                                        var l = now.toLocaleString();
                                        var n = e.get('stats.difficulty');
                                        var shareDiff = e.get('stats.roundShares') / n;
                                        var height = e.get('stats.height');
                                        var point = {
                                            x: now, d: l, y: parseInt(n),
                                            h: height, w: 0, s: shareDiff, f: n
                                        };

                                        var luck = {};
                                        Object.assign(luck, point);
                                        luck.y = shareDiff * 100;

                                        // update chart
                                        series[0].addPoint(point, false, shift);
                                        // add and redraw
                                        series[1].addPoint(luck, true, shift);
                                        return;
                                    }

                                    // new data found, redraw charts using updated data
                                    var diffs = [];
                                    t.forEach(function(d) {
                                        var r = new Date(1000 * d.x);
                                        var l = r.toLocaleString();
                                        var n = d.difficulty;
                                        diffs.push({
                                            x: r, d: l, y: n,
                                            h: d.height, w: d.reward, s: d.sharesDiff, f: d.difficulty
                                        });
                                    });
                                    var lucks = [];
                                    t.forEach(function(d) {
                                        var r = new Date(1000 * d.x);
                                        var l = r.toLocaleString();
                                        var n = d.sharesDiff * 100;
                                        lucks.push({
                                            x: r, d: l, y: n,
                                            h: d.height, w: d.reward, s: d.sharesDiff, f: d.difficulty
                                        });
                                    });
                                    if (diffs.length > 0) {
                                        series[0].setData(diffs, true, {}, true);
                                        e.set('lastData', diffs[diffs.length - 1]); // update lastData
                                    }
                                    if (lucks.length > 0) {
                                        series[1].setData(lucks, true, {}, true);
                                    }
                                }, e.getWithDefault('config.highcharts.blocks.interval', 10000));
                            }
                        }
                    },
                    title: {
                        text: ""
                    },
                    xAxis: {
                        labels: {
                            style: {
                                color: '#6e6e70'
                            }
                        },
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
                    yAxis: [{
                        labels: {
                            style: {
                                color: '#6e6e70'
                            },
                            formatter: function() {
                                var f = this.value;
                                var units = ['H', 'KH', 'MH', 'GH', 'TH'];
                                for (var i = 0; i < 5 && f > 1000; i++)  {
                                    f /= 1000;
                                }
                                return f.toFixed(2) + ' ' + units[i];
                            }
                        },
                        title: {
                            text: "Difficulty",
                            style: {
                                color: 'black',
                                fontSize: '13px',
                                fontWeight: 'normal'
                            }
                        },
                        softMax: 100,
                        gridLineColor: "#e6e6e6"
                    }, {
                        labels: {
                            style: {
                                color: '#6e6e70'
                            },
                            formatter: function() {
                                return this.value.toFixed(0) + ' %';
                            }
                        },
                        title: {
                            text: "Luck",
                            style: {
                                color: 'black',
                                fontSize: '13px',
                                fontWeight: 'normal'
                            }
                        },
                        opposite: true,
                        softMax: 100,
                        gridLineColor: "#e6e6e6",

                        plotLines: [{
                            value: 100,
                            width: 2,
                            color: "#4398de",
                            label: {
                            text: 'expected: 100 %',
                            align: 'center',
                            style: {
                              color: 'gray'
                            }
                          }
                       }]
                    }],
                    plotOptions: {
                        series: {
                            shadow: true
                        },
                        candlestick: {
                            lineColor: '#404048'
                        },
                        map: {
                            shadow: false
                        }
                    },
                    legend: {
                        enabled: true,
                        itemStyle: {}
                    },
                    tooltip: {
                        formatter: function() {
                            function scale(v) {
                                var f = v;
                                var units = ['H', 'KH', 'MH', 'GH', 'TH'];
                                for (var i = 0; i < 5 && f > 1000; i++)  {
                                    f /= 1000;
                                }
                                return f.toFixed(2) + ' ' + units[i];
                            }

                            var d = scale(this.point.f);
                            var s = scale(this.point.s);

                            return "<div>" +
                              "<b>Difficulty: " + d + "</b><br/>" +
                              "<b>Shares: " + s + "</b><br/>" +
                              "<b>Luck: " + (this.point.s*100).toFixed(2)+ " %</b><br/>" +
                              (this.point.w > 0 ? "<b>Reward:&nbsp;" + (this.point.w/1000000000000000000).toFixed(6) + ' ' + e.get('config.Unit') + "</b><br/>" : '') +
                              "<b>Block Height: #" + this.point.h + "</b><br/>" +
                              "<b>" + this.point.d + "</b><br/>" +
                              "</div>";
                        },

                        useHTML: true
                    },
                    exporting: {
                        enabled: true
                    },
                    series: [{
                        yAxis: 0,
                        step: 'center',
                        color: "#E99002",
                        name: "difficulty",
                        data: function() {
                            var a = [];
                            if (null != t) {
                                t.forEach(function(i) {
                                    var n = 0, r = 0, l = 0;
                                    r = new Date(1e3 * i.x);
                                    l = r.toLocaleString();
                                    n = i.difficulty;
                                    a.push({
                                        x: r, d: l, y: n,
                                        h: i.height, w: i.reward, s: i.sharesDiff, f: i.difficulty
                                    });
                                });
                            } else {
                                a.push({ x: 0, d: 0, y: 0, h: 0, w: 0, s: 0, f: 0 });
                            }
                            return a;
                        }()
                    }, {
                        yAxis: 1,
                        step: 'center',
                        color: "#3db72f",
                        name: "Luck",
                        data: function() {
                            var a = [];
                            if (null != t) {
                                t.forEach(function(i) {
                                    var n = 0, r = 0, l = 0;
                                    r = new Date(1e3 * i.x);
                                    l = r.toLocaleString();
                                    n = i.sharesDiff * 100;
                                    a.push({
                                        x: r, d: l, y: n,
                                        h: i.height, w: i.reward, s: i.sharesDiff, f: i.difficulty
                                    });
                                });
                            } else {
                                a.push({ x: 0, d: 0, y: 0, h: 0, w: 0, s: 0, f: 0 });
                            }
                            return a;
                        }()
                    }]
                };
            a.title.text = this.getWithDefault('config.highcharts.blocks.title', '');
            a.chart.height = this.getWithDefault('config.highcharts.blocks.height', 200);

            a.chart.backgroundColor = this.getWithDefault('config.highcharts.blocks.backgroundColor', "transparent");
            a.xAxis.lineColor = this.getWithDefault('config.highcharts.blocks.lineColor', "#ccd6eb");
            a.yAxis.lineColor = this.getWithDefault('config.highcharts.blocks.lineColor', "#ccd6eb");
            a.xAxis.tickColor = this.getWithDefault('config.highcharts.blocks.tickColor', "#ccd6eb");
            a.yAxis.tickColor = this.getWithDefault('config.highcharts.blocks.tickColor', "#ccd6eb");
            a.xAxis.gridLineColor = this.getWithDefault('config.highcharts.blocks.gridLineColor', "#ccd6eb");
            a.xAxis.gridLineWidth = this.getWithDefault('config.highcharts.blocks.gridLineWidthX', "0");
            a.yAxis[0].gridLineColor = this.getWithDefault('config.highcharts.blocks.gridLineColor', "#ccd6eb");
            a.yAxis[1].gridLineColor = this.getWithDefault('config.highcharts.blocks.gridLineColor', "#ccd6eb");
            a.yAxis[0].title.style.color = this.getWithDefault('config.highcharts.blocks.labelColor', 'black');
            a.yAxis[1].title.style.color = this.getWithDefault('config.highcharts.blocks.labelColor', 'black');
            a.yAxis[1].plotLines[0].color = this.getWithDefault('config.highcharts.blocks.plotLineColor', '#4398de');
            a.legend.itemStyle.color = this.getWithDefault('config.highcharts.blocks.labelColor', "#fff");
            return a;
        }
    })

});
