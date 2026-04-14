import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from 'recharts';

interface DataPoint {
  timestamp: string;
  value: number;
}

interface HealthChartProps {
  data: DataPoint[];
  label: string;
  unit: string;
  color?: string;
}

function formatTime(ts: string): string {
  const d = new Date(ts);
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`;
}

export default function HealthChart({
  data,
  label,
  unit,
  color = '#3b82f6',
}: HealthChartProps) {
  return (
    <div className="w-full h-64">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 5, right: 20, left: 10, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" className="opacity-30" />
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatTime}
            tick={{ fontSize: 11 }}
            interval="preserveStartEnd"
          />
          <YAxis
            tick={{ fontSize: 11 }}
            label={{
              value: `${label} (${unit})`,
              angle: -90,
              position: 'insideLeft',
              style: { fontSize: 12 },
            }}
          />
          <Tooltip
            labelFormatter={(ts) => new Date(ts as string).toLocaleString()}
            formatter={(val: number) => [`${val} ${unit}`, label]}
          />
          <Line
            type="monotone"
            dataKey="value"
            stroke={color}
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
