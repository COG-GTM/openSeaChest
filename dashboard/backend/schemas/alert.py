"""Request/response schemas for alert endpoints."""

from datetime import datetime
from typing import Optional

from pydantic import BaseModel


class AlertRuleCreate(BaseModel):
    name: str
    description: Optional[str] = None
    rule_type: str
    threshold_value: Optional[float] = None
    approved_firmware_list: Optional[list[str]] = None
    enabled: bool = True


class AlertRuleUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    rule_type: Optional[str] = None
    threshold_value: Optional[float] = None
    approved_firmware_list: Optional[list[str]] = None
    enabled: Optional[bool] = None


class AlertRuleResponse(BaseModel):
    id: int
    name: str
    description: Optional[str] = None
    rule_type: str
    threshold_value: Optional[float] = None
    approved_firmware_list: Optional[list[str]] = None
    enabled: bool
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


class FiredAlertResponse(BaseModel):
    id: int
    rule_id: int
    device_id: int
    device_serial: Optional[str] = None
    rule_name: Optional[str] = None
    fired_at: datetime
    message: Optional[str] = None
    resolved: bool
    resolved_at: Optional[datetime] = None

    model_config = {"from_attributes": True}


class FiredAlertListResponse(BaseModel):
    total: int
    alerts: list[FiredAlertResponse]
