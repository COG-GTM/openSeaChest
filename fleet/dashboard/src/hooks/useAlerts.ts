import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAlerts, acknowledgeAlert, getAlertRules } from '../api/client';
import type { AlertQueryParams } from '../api/types';

export function useAlerts(params?: AlertQueryParams) {
  return useQuery({
    queryKey: ['alerts', params],
    queryFn: () => getAlerts(params),
    refetchInterval: 30000,
    placeholderData: (prev) => prev,
  });
}

export function useAlertRules() {
  return useQuery({
    queryKey: ['alert-rules'],
    queryFn: getAlertRules,
  });
}

export function useAcknowledgeAlert() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, annotation }: { id: string; annotation: string }) =>
      acknowledgeAlert(id, annotation),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
    },
  });
}
