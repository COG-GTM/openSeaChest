import { getStatusColor } from '../theme';

interface StatusBadgeProps {
  status: string;
  size?: 'sm' | 'md' | 'lg';
}

const SIZE_CLASSES = {
  sm: 'px-1.5 py-0.5 text-xs',
  md: 'px-2 py-1 text-xs',
  lg: 'px-3 py-1 text-sm',
} as const;

const STATUS_LABELS: Record<string, string> = {
  good: 'Good',
  warning: 'Warning',
  tripped: 'Tripped',
  unknown: 'Unknown',
  compliant: 'Compliant',
  non_compliant: 'Non-Compliant',
  no_policy: 'No Policy',
  firing: 'Firing',
  acknowledged: 'Acknowledged',
  resolved: 'Resolved',
  critical: 'Critical',
  info: 'Info',
};

export default function StatusBadge({ status, size = 'md' }: StatusBadgeProps) {
  const color = getStatusColor(status);
  const label = STATUS_LABELS[status] ?? status;

  return (
    <span
      className={`inline-flex items-center font-medium rounded-full ${SIZE_CLASSES[size]}`}
      style={{
        backgroundColor: `${color}20`,
        color: color,
        border: `1px solid ${color}40`,
      }}
    >
      <span
        className="w-1.5 h-1.5 rounded-full mr-1.5"
        style={{ backgroundColor: color }}
      />
      {label}
    </span>
  );
}
