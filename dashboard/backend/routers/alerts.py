"""Alert rules CRUD and fired alerts listing."""

from datetime import datetime, timezone
from typing import Optional

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..database import get_db
from ..models.alert_rule import AlertRule, RuleType
from ..models.fired_alert import FiredAlert
from ..models.device import Device
from ..schemas.alert import (
    AlertRuleCreate,
    AlertRuleUpdate,
    AlertRuleResponse,
    FiredAlertResponse,
    FiredAlertListResponse,
)

router = APIRouter()


# --- Alert Rules CRUD ---


@router.post("/alerts/rules", response_model=AlertRuleResponse, status_code=201)
def create_alert_rule(rule_in: AlertRuleCreate, db: Session = Depends(get_db)):
    """Create a new alert rule."""
    try:
        rule_type = RuleType(rule_in.rule_type)
    except ValueError:
        raise HTTPException(
            status_code=422,
            detail=f"Invalid rule_type: {rule_in.rule_type}. "
            f"Must be one of: {[r.value for r in RuleType]}",
        )

    now = datetime.now(timezone.utc)
    rule = AlertRule(
        name=rule_in.name,
        description=rule_in.description,
        rule_type=rule_type,
        threshold_value=rule_in.threshold_value,
        approved_firmware_list=rule_in.approved_firmware_list,
        enabled=rule_in.enabled,
        created_at=now,
        updated_at=now,
    )
    db.add(rule)
    db.commit()
    db.refresh(rule)
    return _rule_to_response(rule)


@router.get("/alerts/rules", response_model=list[AlertRuleResponse])
def list_alert_rules(db: Session = Depends(get_db)):
    """List all alert rules."""
    rules = db.query(AlertRule).all()
    return [_rule_to_response(r) for r in rules]


@router.get("/alerts/rules/{rule_id}", response_model=AlertRuleResponse)
def get_alert_rule(rule_id: int, db: Session = Depends(get_db)):
    """Get alert rule by ID."""
    rule = db.query(AlertRule).filter(AlertRule.id == rule_id).first()
    if not rule:
        raise HTTPException(status_code=404, detail="Alert rule not found")
    return _rule_to_response(rule)


@router.put("/alerts/rules/{rule_id}", response_model=AlertRuleResponse)
def update_alert_rule(
    rule_id: int, rule_in: AlertRuleUpdate, db: Session = Depends(get_db)
):
    """Update an alert rule."""
    rule = db.query(AlertRule).filter(AlertRule.id == rule_id).first()
    if not rule:
        raise HTTPException(status_code=404, detail="Alert rule not found")

    update_data = rule_in.model_dump(exclude_unset=True)

    if "rule_type" in update_data:
        if update_data["rule_type"] is None:
            del update_data["rule_type"]
        else:
            try:
                update_data["rule_type"] = RuleType(update_data["rule_type"])
            except ValueError:
                raise HTTPException(
                    status_code=422,
                    detail=f"Invalid rule_type: {update_data['rule_type']}",
                )

    for field, value in update_data.items():
        setattr(rule, field, value)

    rule.updated_at = datetime.now(timezone.utc)
    db.commit()
    db.refresh(rule)
    return _rule_to_response(rule)


@router.delete("/alerts/rules/{rule_id}", status_code=204)
def delete_alert_rule(rule_id: int, db: Session = Depends(get_db)):
    """Delete an alert rule."""
    rule = db.query(AlertRule).filter(AlertRule.id == rule_id).first()
    if not rule:
        raise HTTPException(status_code=404, detail="Alert rule not found")

    # Delete associated fired alerts first
    db.query(FiredAlert).filter(FiredAlert.rule_id == rule_id).delete()
    db.delete(rule)
    db.commit()
    return None


# --- Fired Alerts ---


@router.get("/alerts", response_model=FiredAlertListResponse)
def list_fired_alerts(
    device_serial: Optional[str] = Query(None),
    rule_id: Optional[int] = Query(None),
    resolved: Optional[bool] = Query(None),
    db: Session = Depends(get_db),
):
    """List fired alerts with optional filters."""
    query = db.query(FiredAlert)

    if device_serial:
        device = db.query(Device).filter(Device.serial_number == device_serial).first()
        if device:
            query = query.filter(FiredAlert.device_id == device.id)
        else:
            return FiredAlertListResponse(total=0, alerts=[])

    if rule_id is not None:
        query = query.filter(FiredAlert.rule_id == rule_id)

    if resolved is not None:
        query = query.filter(FiredAlert.resolved == resolved)

    alerts = query.order_by(FiredAlert.fired_at.desc()).all()

    return FiredAlertListResponse(
        total=len(alerts),
        alerts=[_fired_alert_to_response(a, db) for a in alerts],
    )


# --- Helpers ---


def _rule_to_response(rule: AlertRule) -> AlertRuleResponse:
    return AlertRuleResponse(
        id=rule.id,
        name=rule.name,
        description=rule.description,
        rule_type=rule.rule_type.value,
        threshold_value=rule.threshold_value,
        approved_firmware_list=rule.approved_firmware_list,
        enabled=rule.enabled,
        created_at=rule.created_at,
        updated_at=rule.updated_at,
    )


def _fired_alert_to_response(alert: FiredAlert, db: Session) -> FiredAlertResponse:
    device = db.query(Device).filter(Device.id == alert.device_id).first()
    rule = db.query(AlertRule).filter(AlertRule.id == alert.rule_id).first()
    return FiredAlertResponse(
        id=alert.id,
        rule_id=alert.rule_id,
        device_id=alert.device_id,
        device_serial=device.serial_number if device else None,
        rule_name=rule.name if rule else None,
        fired_at=alert.fired_at,
        message=alert.message,
        resolved=alert.resolved,
        resolved_at=alert.resolved_at,
    )
