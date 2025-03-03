// Format and extract units for display
function extractUnit(value) {
    const strValue = String(value);
    const match = strValue.match(/[°a-zA-Z%]+/);
    return match ? match[0] : '';
}

function formatValue(value) {
    if (value === undefined || value === null) {
        return { value: 0, unit: '' };
    }
    const numericValue = parseFloat(value);
    const unit = extractUnit(value);
    return { value: numericValue, unit: unit };
}

// Format and display values in stat cards
function updateStatCards(data) {
    // CPU usage
    if (data.cpu_usage) {
        document.getElementById('cpuUsage').textContent = data.cpu_usage;
    }

    // Memory usage
    if (data.used_ram_percentage) {
        document.getElementById('memoryUsage').textContent = data.used_ram_percentage;
    }

    // Disk usage
    if (data.disk_usage_percent) {
        document.getElementById('diskUsage').textContent = data.disk_usage_percent;
    }

    // Uptime
    if (data.uptime) {
        document.getElementById('uptime').textContent = data.uptime;
    }

    // Load average
    if (data.load_1 && data.load_5 && data.load_15) {
        document.getElementById('loadAvg').textContent =
            `${data.load_1} | ${data.load_5} | ${data.load_15}`;
    }

    // Process count
    if (data.process_count) {
        document.getElementById('processCount').textContent = data.process_count;
    }

    // Swap memory
    if (data.swap_percent) {
        document.getElementById('swapUsage').textContent = data.swap_percent;
    }

    // CPU frequency
    if (data.cpu_mhz) {
        document.getElementById('cpuFreq').textContent = data.cpu_mhz;
    }
}

// Initialize all charts
function initializeCharts() {
    // Performance chart
    const performanceChart = echarts.init(document.getElementById('chart'));

    // Storage chart
    const storageChartDom = document.getElementById('storageChart');
    const storageChart = echarts.init(storageChartDom);

    // Network chart
    const networkChartDom = document.getElementById('networkChart');
    const networkChart = echarts.init(networkChartDom);

    // Disk chart
    const diskChartDom = document.getElementById('diskChart');
    const diskChart = echarts.init(diskChartDom);

    return {
        performance: performanceChart,
        storage: storageChart,
        network: networkChart,
        disk: diskChart
    };
}

// Set up the WebSocket connection and handle data
function webSocket() {
    const clientID = window.clientID;
    const host = window.location.hostname;
    const port = window.location.port;
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';

    // Show loading state
    document.getElementById('loading').style.display = 'block';
    document.getElementById('error').style.display = 'none';

    const ws = new WebSocket(`${protocol}://${host}:${port}/analytics_ws/${clientID}`);

    // Initialize all charts
    const charts = initializeCharts();
    const chartDevice = charts.performance;

    const option = {
        tooltip: {
            trigger: 'axis'
        },
        legend: {
            type: 'scroll',
            data: ['Average Chipset Temp', 'CPU Temp', 'CPU Usage', 'Free RAM', 'Used RAM', 'Used RAM Percentage', 'Packets Received', 'Packets Sent'],
            selected: {
                'Packets Received': false,
                'Packets Sent': false,
                'Free RAM': false,
            }
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: []
        },
        yAxis: {
            type: 'value',
        },
        series: [
            { name: 'Average Chipset Temp', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'CPU Temp', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'CPU Usage', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Free RAM', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Used RAM', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Used RAM Percentage', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Packets Received', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Packets Sent', type: 'line', data: [], markPoint: { data: [] } }
        ]
    };

    // Initialize network chart
    const networkOption = {
        tooltip: { trigger: 'axis' },
        legend: { data: ['Packets Received', 'Packets Sent'] },
        xAxis: { type: 'category', boundaryGap: false, data: [] },
        yAxis: { type: 'value' },
        series: [
            { name: 'Packets Received', type: 'line', data: [], areaStyle: {}, smooth: true },
            { name: 'Packets Sent', type: 'line', data: [], areaStyle: {}, smooth: true }
        ]
    };

    // Initialize storage chart
    const storageOption = {
        tooltip: { trigger: 'axis' },
        legend: { data: ['Free RAM', 'Used RAM', 'Used RAM Percentage', 'Swap Used'] },
        xAxis: { type: 'category', boundaryGap: false, data: [] },
        yAxis: { type: 'value' },
        series: [
            { name: 'Free RAM', type: 'line', data: [], smooth: true },
            { name: 'Used RAM', type: 'line', data: [], smooth: true },
            { name: 'Used RAM Percentage', type: 'line', data: [], smooth: true },
            { name: 'Swap Used', type: 'line', data: [], smooth: true }
        ]
    };

    // Initialize disk chart
    const diskOption = {
        tooltip: {
            trigger: 'item',
            formatter: '{b}: {c} ({d}%)'
        },
        legend: {
            orient: 'vertical',
            left: 'left',
            data: ['Used Space', 'Free Space']
        },
        series: [
            {
                name: 'Disk Usage',
                type: 'pie',
                radius: '70%',
                center: ['50%', '50%'],
                data: [
                    {value: 0, name: 'Used Space'},
                    {value: 0, name: 'Free Space'}
                ],
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)'
                    }
                },
                label: {
                    formatter: '{b}: {c} ({d}%)'
                }
            }
        ]
    };

    chartDevice.setOption(option);
    charts.network.setOption(networkOption);
    charts.storage.setOption(storageOption);
    charts.disk.setOption(diskOption);

    // Keep track of legend selection
    let currentLegend = chartDevice.getOption().legend[0].selected;
    chartDevice.on('legendselectchanged', function(params) {
        currentLegend = { ...params.selected };
    });

    ws.onmessage = function(event) {
        // Hide loading state
        document.getElementById('loading').style.display = 'none';

        const data = JSON.parse(event.data);
        const time = new Date().toLocaleTimeString();
        //console.log(data);

        // Update time for all charts
        option.xAxis.data.push(time);
        networkOption.xAxis.data.push(time);
        storageOption.xAxis.data.push(time);

        // Limit data points for better performance
        if (option.xAxis.data.length > 500) {
            option.xAxis.data.shift();
            networkOption.xAxis.data.shift();
            storageOption.xAxis.data.shift();
        }

        const seriesData = [
            formatValue(data.average_chipset_temp),
            formatValue(data.cpu_temp),
            formatValue(data.cpu_usage),
            formatValue(data.free_ram),
            formatValue(data.used_ram),
            formatValue(data.used_ram_percentage),
            formatValue(data.packets_receive),
            formatValue(data.packets_sent)
        ];

        // Update performance chart data
        seriesData.forEach((item, index) => {
            option.series[index].data.push(item);
            if (option.series[index].data.length > 500) {
                option.series[index].data.shift();
            }

            const values = option.series[index].data.map(d => d.value);
            const maxValue = Math.max(...values);
            const minValue = Math.min(...values);

            option.series[index].markPoint.data = [
                { type: 'max', name: 'Max', value: maxValue, label: { formatter: `{c} ${item.unit}` } },
                { type: 'min', name: 'Min', value: minValue, label: { formatter: `{c} ${item.unit}` } }
            ];
        });

        // Update network chart data
        networkOption.series[0].data.push(seriesData[6]);
        networkOption.series[1].data.push(seriesData[7]);

        if (networkOption.series[0].data.length > 500) {
            networkOption.series[0].data.shift();
            networkOption.series[1].data.shift();
        }

        // Update storage chart data correctly
        if (data.free_ram) storageOption.series[0].data.push(seriesData[3].value);
        if (data.used_ram) storageOption.series[1].data.push(seriesData[4].value);
        if (data.used_ram_percentage) storageOption.series[2].data.push(seriesData[5].value);

        // Add swap data if available
        if (data.swap_used) {
            const swapData = formatValue(data.swap_used);
            storageOption.series[3].data.push(swapData.value);

            if (storageOption.series[3].data.length > 500) {
                storageOption.series[3].data.shift();
            }
        }

        if (storageOption.series[0].data.length > 500) {
            storageOption.series[0].data.shift();
            storageOption.series[1].data.shift();
            storageOption.series[2].data.shift();
        }

        // Update disk chart if disk data is available
        if (data.disk_used && data.disk_free) {
            const usedData = formatValue(data.disk_used);
            const freeData = formatValue(data.disk_free);
            const diskUnit = usedData.unit || freeData.unit || '';

            diskOption.series[0].data = [
                {value: usedData.value, name: 'Used Space', unit: diskUnit},
                {value: freeData.value, name: 'Free Space', unit: diskUnit}
            ];

            // Update tooltip and label formatter to include the unit
            diskOption.tooltip.formatter = function(params) {
                return `${params.name}: ${params.value} ${params.data.unit} (${params.percent}%)`;
            };

            diskOption.series[0].label.formatter = function(params) {
                return `${params.name}: ${params.value} ${params.data.unit} (${params.percent}%)`;
            };
        }

        // Configure tooltips
        const tooltipFormatter = function(params) {
            let result = params[0].axisValue + '<br/>';
            params.forEach(item => {
                if (item.data && typeof item.data === 'object') {
                    result += item.marker + item.seriesName + ': ' + item.data.value + ' ' + item.data.unit + '<br/>';
                } else {
                    result += item.marker + item.seriesName + ': ' + item.data + '<br/>';
                }
            });
            return result;
        };

        option.tooltip.formatter = tooltipFormatter;
        networkOption.tooltip.formatter = tooltipFormatter;
        storageOption.tooltip.formatter = tooltipFormatter;

        option.legend.selected = currentLegend;

        // Update all charts with new options
        chartDevice.setOption(option);
        charts.network.setOption(networkOption);
        charts.storage.setOption(storageOption);
        charts.disk.setOption(diskOption);

        updateStatCards(data);
    };

    ws.onopen = function() {
        console.log("Connected to WebSocket");
        document.getElementById('loading').style.display = 'none';
        document.getElementById('error').style.display = 'none';
    };

    ws.onclose = function() {
        console.log("Connection lost, retrying...");
        document.getElementById('loading').style.display = 'none';
        document.getElementById('error').style.display = 'block';
        setTimeout(webSocket, 3000);
    };

    ws.onerror = function(error) {
        console.log("WebSocket Error:", error);
        document.getElementById('loading').style.display = 'none';
        document.getElementById('error').style.display = 'block';
        ws.close();
    };

    // Handle the retry button click
    document.getElementById('retryButton').addEventListener('click', function() {
        document.getElementById('error').style.display = 'none';
        document.getElementById('loading').style.display = 'block';
        webSocket();
    });

    // Handle filter apply button
    // document.getElementById('applyFilters').addEventListener('click', function() {
    //     const metric = document.getElementById('metricSelect').value;
    //     const timeRange = document.getElementById('timeRange').value;
    //
    //     // Update chart visibility based on metric selection
    //     if (metric === 'cpu') {
    //         currentLegend = {
    //             'Average Chipset Temp': true,
    //             'CPU Temp': true,
    //             'CPU Usage': true,
    //             'Free RAM': false,
    //             'Used RAM': false,
    //             'Used RAM Percentage': false,
    //             'Packets Received': false,
    //             'Packets Sent': false
    //         };
    //     } else if (metric === 'memory') {
    //         currentLegend = {
    //             'Average Chipset Temp': false,
    //             'CPU Temp': false,
    //             'CPU Usage': false,
    //             'Free RAM': true,
    //             'Used RAM': true,
    //             'Used RAM Percentage': true,
    //             'Packets Received': false,
    //             'Packets Sent': false
    //         };
    //     } else if (metric === 'disk') {
    //         // Switch to the disk tab
    //         document.getElementById('disk-tab').click();
    //         return;
    //     } else if (metric === 'all') {
    //         currentLegend = {
    //             'Average Chipset Temp': true,
    //             'CPU Temp': true,
    //             'CPU Usage': true,
    //             'Free RAM': true,
    //             'Used RAM': true,
    //             'Used RAM Percentage': true,
    //             'Packets Received': false,
    //             'Packets Sent': false
    //         };
    //     }
    //
    //     option.legend.selected = currentLegend;
    //     chartDevice.setOption(option);
    // });

    // Handle tab changes to resize charts properly
    const tabElements = document.querySelectorAll('button[data-bs-toggle="tab"]');
    tabElements.forEach(function(tabElement) {
        tabElement.addEventListener('shown.bs.tab', function() {
            charts.performance.resize();
            charts.network.resize();
            charts.storage.resize();
            charts.disk.resize();
        });
    });

    // Handle window resize
    window.addEventListener('resize', function() {
        charts.performance.resize();
        charts.network.resize();
        charts.storage.resize();
        charts.disk.resize();
    });
}

// Set initial chart dimensions
function setChartDimensions() {
    const chartDivs = ['chart', 'storageChart', 'networkChart', 'diskChart'];
    const screenHeight = window.innerHeight;

    chartDivs.forEach(id => {
        const chartDiv = document.getElementById(id);
        if (chartDiv) {
            chartDiv.style.height = Math.max(400, screenHeight * 0.5) + 'px';
        }
    });
}

// Initialize everything when the page loads
document.addEventListener('DOMContentLoaded', function() {
    setChartDimensions();
    webSocket();

    // Set current date for date pickers
    const today = new Date();
    const lastWeek = new Date(today);
    lastWeek.setDate(today.getDate() - 7);

    // document.getElementById('endDate').valueAsDate = today;
    // document.getElementById('startDate').valueAsDate = lastWeek;
});