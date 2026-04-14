export const COLORS = {
  good: '#22c55e',
  warning: '#f59e0b',
  tripped: '#ef4444',
  unknown: '#9ca3af',
  compliant: '#22c55e',
  non_compliant: '#ef4444',
  no_policy: '#9ca3af',
  primary: '#3b82f6',
  info: '#3b82f6',
  critical: '#ef4444',
} as const;

export const CHART_COLORS = [
  '#3b82f6',
  '#22c55e',
  '#f59e0b',
  '#ef4444',
  '#8b5cf6',
  '#ec4899',
  '#06b6d4',
  '#84cc16',
] as const;

export function getStatusColor(
  status: string,
): string {
  return COLORS[status as keyof typeof COLORS] ?? COLORS.unknown;
}

export function initTheme(): void {
  const stored = localStorage.getItem('fleet_theme');
  if (stored === 'dark' || (!stored && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
}

export function toggleTheme(): void {
  const isDark = document.documentElement.classList.toggle('dark');
  localStorage.setItem('fleet_theme', isDark ? 'dark' : 'light');
}

export function isDarkMode(): boolean {
  return document.documentElement.classList.contains('dark');
}
