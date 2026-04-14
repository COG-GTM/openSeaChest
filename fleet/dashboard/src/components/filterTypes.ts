import type { InterfaceType, SmartStatus, ComplianceStatus } from '../api/types';

export interface FilterValues {
  model: string;
  firmware: string;
  interface_type: InterfaceType | '';
  host: string;
  health_status: SmartStatus | '';
  compliance_status: ComplianceStatus | '';
}

export const EMPTY_FILTERS: FilterValues = {
  model: '',
  firmware: '',
  interface_type: '',
  host: '',
  health_status: '',
  compliance_status: '',
};
