import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from "recharts";

interface CompliancePieChartProps {
  compliantCount: number;
  nonCompliantCount: number;
}

const COLORS = ["#22c55e", "#ef4444"];

export default function CompliancePieChart({
  compliantCount,
  nonCompliantCount,
}: CompliancePieChartProps) {
  const data = [
    { name: "Compliant", value: compliantCount },
    { name: "Non-Compliant", value: nonCompliantCount },
  ];

  if (compliantCount === 0 && nonCompliantCount === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-400">
        No data to display
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={300}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          labelLine={false}
          label={({ name, percent }) =>
            `${name}: ${(percent * 100).toFixed(0)}%`
          }
          outerRadius={100}
          fill="#8884d8"
          dataKey="value"
        >
          {data.map((_entry, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  );
}
