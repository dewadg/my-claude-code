"""Turns the scan jobs' outputs into one JUnit report for the pipeline Tests tab.

Usage: python3 .gitlab/scan-junit.py OUTPUT.xml NAME [NAME ...]

Each NAME is a report at scan-reports/NAME.json. A report with a "results" key is
osv-scanner --format json; anything else is read as the GitLab security report schema
(SAST). Every finding becomes a failed test case, an osv advisory whose Go code call
analysis marks as uncalled becomes a skipped one (it does not gate the dependency scan),
and a scanner with no findings gets one passing case. A missing report is an error case,
so a scan that died before writing one still shows up.
"""
import json
import os
import sys
import xml.etree.ElementTree as ET

REPORTS = "scan-reports"


def case(suite, name, kind=None, message="", text=""):
    c = ET.SubElement(suite, "testcase", classname=suite.get("name"), name=name)
    if kind:
        ET.SubElement(c, kind, message=message).text = text


def gitlab(suite, report):
    for v in report["vulnerabilities"]:
        loc = v.get("location", {})
        ids = ", ".join(i["value"] for i in v.get("identifiers", []))
        case(suite, f"{v['severity']}: {v['name']} ({loc.get('file')}:{loc.get('start_line')})",
             "failure", ids, v.get("description", ""))


def osv(suite, report):
    for r in report["results"]:
        lockfile = os.path.relpath(r["source"]["path"], os.environ.get("CI_PROJECT_DIR", "."))
        for p in r["packages"]:
            pkg = p["package"]
            summaries = {v["id"]: v.get("summary", "") for v in p["vulnerabilities"]}
            for g in p["groups"]:
                analysis = g.get("experimental_analysis", {})
                uncalled = analysis and all(not a.get("called", True) or a.get("unimportant") for a in analysis.values())
                severity = g.get("max_severity") or "?"
                name = f"{g['ids'][0]} {pkg['name']}@{pkg['version']} ({lockfile})"
                text = "\n".join(f"{i}: {summaries.get(i, '')}" for i in g["ids"])
                if uncalled:
                    case(suite, name, "skipped", "uncalled, does not gate", text)
                else:
                    case(suite, name, "failure", f"severity {severity}; aliases {', '.join(g['aliases'])}", text)


def main():
    if len(sys.argv) < 3:
        sys.exit(__doc__)
    root = ET.Element("testsuites")
    for name in sys.argv[2:]:
        path = f"{REPORTS}/{name}.json"
        suite = ET.SubElement(root, "testsuite", name=name)
        if not os.path.exists(path):
            case(suite, "report", "error", f"{path} missing: the scan did not finish")
            continue
        report = json.load(open(path))
        osv(suite, report) if "results" in report else gitlab(suite, report)
        if not len(suite):
            case(suite, "no findings")

    for suite in root:
        counts = {k: len(suite.findall(f"testcase/{k}")) for k in ("failure", "skipped", "error")}
        suite.set("tests", str(len(suite)))
        suite.set("failures", str(counts["failure"]))
        suite.set("skipped", str(counts["skipped"]))
        suite.set("errors", str(counts["error"]))
        print(f"{suite.get('name'):<16} findings={counts['failure']} uncalled={counts['skipped']} missing={counts['error']}")
    ET.ElementTree(root).write(sys.argv[1], encoding="utf-8", xml_declaration=True)


main()
