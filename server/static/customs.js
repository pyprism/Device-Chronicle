function extractUnit(value) {
    const strValue = String(value);
    const match = strValue.match(/[°a-zA-Z%]+/);
    return match ? match[0] : '';
}

function formatValue(value) {
    const numericValue = parseFloat(value);
    const unit = extractUnit(value);
    return { value: numericValue, unit: unit };
}

function webSocket() {
    const clientID = window.clientID;
    const host = window.location.hostname;
    const port = window.location.port;
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const ws = new WebSocket(`${protocol}://${host}:${port}/analytics_ws/${clientID}`);

    const chartDom = document.getElementById('chart');
    const chartDevice = echarts.init(chartDom);

    const option = {
        tooltip: {
            trigger: 'axis'
        },
        legend: {
            data: ['Average Chipset Temp', 'CPU Temp', 'Free RAM', 'Used RAM', 'Used RAM Percentage', 'Packets Received', 'Packets Sent'],
            selected: {
                'Packets Received': false,
                'Packets Sent': false
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
            { name: 'Free RAM', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Used RAM', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Used RAM Percentage', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Packets Received', type: 'line', data: [], markPoint: { data: [] } },
            { name: 'Packets Sent', type: 'line', data: [], markPoint: { data: [] } }
        ]
    };

    chartDevice.setOption(option);

    // Keep track of the current legend selection and update it when the legend is clicked
    let currentLegend = chartDevice.getOption().legend[0].selected;
    chartDevice.on('legendselectchanged', function(params) {
        currentLegend = { ...params.selected };
    });


    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        console.log(data)
        const time = new Date().toLocaleTimeString();

        option.xAxis.data.push(time);
        if (option.xAxis.data.length > 500) {
            option.xAxis.data.shift();
        }

        const seriesData = [
            formatValue(data.average_chipset_temp),
            formatValue(data.cpu_temp),
            formatValue(data.free_ram),
            formatValue(data.used_ram),
            formatValue(data.used_ram_percentage),
            formatValue(data.packets_receive),
            formatValue(data.packets_sent)
        ];

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

        option.tooltip = {
            trigger: 'axis',
            formatter: function(params) {
                let result = params[0].axisValue + '<br/>';
                params.forEach(item => {
                    result += item.marker + item.seriesName + ': ' + item.data.value + ' ' + item.data.unit + '<br/>';
                });
                return result;
            }
        };
        option.legend.selected = currentLegend;
        chartDevice.setOption(option);
    };

    ws.onopen = function() {
        console.log("Connected to WebSocket");
    };

    ws.onclose = function() {
        console.log("Connection lost, retrying...");
        setTimeout(webSocket, 3000);
    };

    ws.onerror = function(error) {
        console.log("WebSocket Error:", error);
        ws.close();
    };

    window.addEventListener('resize', function() {
        chartDevice.resize();
    });
}


function setChartDimensions() {
    const chartDiv = document.getElementById('chart');
    const screenHeight = window.innerHeight;
    const screenWidth = window.innerWidth;
    chartDiv.style.width = (screenWidth * 0.9) + 'px'; // Set width to 90% of the screen width
    chartDiv.style.height = (screenHeight * 0.9) + 'px'; // Set height to 70% of the screen height
}

window.addEventListener('resize', setChartDimensions);


document.addEventListener('DOMContentLoaded', function() {
    setChartDimensions();
    webSocket();
});