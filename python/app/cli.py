from __future__ import annotations

import argparse
import sys
import time
from typing import Any

import requests


def main() -> None:
    parser = argparse.ArgumentParser(description="Vitals CLI")
    subparsers = parser.add_subparsers(dest="command", required=True)

    common = argparse.ArgumentParser(add_help=False)
    common.add_argument("--addr", default="127.0.0.1:5000", help="server host:port")

    insert_parser = subparsers.add_parser("insert-vital", parents=[common], help="insert a vital")
    insert_parser.add_argument("--patient", required=True, help="patient identifier")
    insert_parser.add_argument("--systolic", type=int, required=True, help="systolic blood pressure")
    insert_parser.add_argument("--diastolic", type=int, required=True, help="diastolic blood pressure")
    insert_parser.add_argument("--taken-at", type=int, default=0, help="unix timestamp when blood pressure was taken")

    list_alerts_parser = subparsers.add_parser("list-alerts", parents=[common], help="list alerts")
    list_alerts_parser.add_argument("--patient", default="", help="patient identifier")

    list_vitals_parser = subparsers.add_parser("list-vitals", parents=[common], help="list vitals")
    list_vitals_parser.add_argument("--patient", default="", help="patient identifier")

    args = parser.parse_args()

    if args.command == "insert-vital":
        insert_vital_cmd(args)
    elif args.command == "list-alerts":
        list_alerts_cmd(args)
    elif args.command == "list-vitals":
        list_vitals_cmd(args)


def insert_vital_cmd(args: argparse.Namespace) -> None:
    base_url = _base_url(args.addr)
    taken_at = args.taken_at or int(time.time())
    payload = {
        "patient_id": args.patient,
        "systolic": args.systolic,
        "diastolic": args.diastolic,
        "taken_at": taken_at,
    }

    try:
        resp = requests.post(f"{base_url}/vitals", json=payload, timeout=5)
    except requests.RequestException as exc:
        _fail(f"insert vital failed: {exc}")

    if resp.status_code >= 400:
        _fail(f"insert vital failed: {_extract_error(resp)}")

    vital = resp.json().get("vital", {})
    print(
        "stored vital id={id} patient={patient} bp={systolic}/{diastolic} taken_at={taken_at} received_at={received_at}".format(
            id=vital.get("id", 0),
            patient=vital.get("patient_id", ""),
            systolic=vital.get("systolic", 0),
            diastolic=vital.get("diastolic", 0),
            taken_at=vital.get("taken_at", 0),
            received_at=vital.get("received_at", 0),
        )
    )


def list_alerts_cmd(args: argparse.Namespace) -> None:
    base_url = _base_url(args.addr)
    params = {}
    if args.patient:
        params["patient_id"] = args.patient

    try:
        resp = requests.get(f"{base_url}/alerts", params=params, timeout=5)
    except requests.RequestException as exc:
        _fail(f"list alerts failed: {exc}")

    if resp.status_code >= 400:
        _fail(f"list alerts failed: {_extract_error(resp)}")

    alerts = resp.json().get("alerts", [])
    if not alerts:
        print("no alerts")
        return

    for alert in alerts:
        vital = alert.get("vital", {})
        print(
            "alert id={id} patient={patient} bp={systolic}/{diastolic} status={status} reason={reason} created_at={created_at}".format(
                id=alert.get("id", 0),
                patient=vital.get("patient_id", ""),
                systolic=vital.get("systolic", 0),
                diastolic=vital.get("diastolic", 0),
                status=alert.get("status", ""),
                reason=alert.get("reason", ""),
                created_at=alert.get("created_at", 0),
            )
        )


def list_vitals_cmd(args: argparse.Namespace) -> None:
    base_url = _base_url(args.addr)
    params = {}
    if args.patient:
        params["patient_id"] = args.patient

    try:
        resp = requests.get(f"{base_url}/vitals", params=params, timeout=5)
    except requests.RequestException as exc:
        _fail(f"list vitals failed: {exc}")

    if resp.status_code >= 400:
        _fail(f"list vitals failed: {_extract_error(resp)}")

    vitals = resp.json().get("vitals", [])
    if not vitals:
        print("no vitals")
        return

    for vital in vitals:
        print(
            "vital id={id} patient={patient} bp={systolic}/{diastolic} taken_at={taken_at} received_at={received_at}".format(
                id=vital.get("id", 0),
                patient=vital.get("patient_id", ""),
                systolic=vital.get("systolic", 0),
                diastolic=vital.get("diastolic", 0),
                taken_at=vital.get("taken_at", 0),
                received_at=vital.get("received_at", 0),
            )
        )


def _base_url(addr: str) -> str:
    if addr.startswith("http://") or addr.startswith("https://"):
        return addr.rstrip("/")
    return f"http://{addr}"


def _extract_error(resp: requests.Response) -> str:
    try:
        data: dict[str, Any] = resp.json()
    except ValueError:
        return f"status {resp.status_code}"
    return data.get("error", f"status {resp.status_code}")


def _fail(message: str) -> None:
    print(message, file=sys.stderr)
    sys.exit(1)


if __name__ == "__main__":
    main()
