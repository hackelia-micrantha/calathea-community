#!/usr/bin/env python3
from __future__ import annotations

import argparse
import datetime as dt
import gzip
import hashlib
import io
import json
import re
import tarfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SEMVER = re.compile(r"^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z.-]+))?(?:\+([0-9A-Za-z.-]+))?$")


def version() -> str:
    value = (ROOT / "VERSION").read_text(encoding="utf-8").strip()
    if SEMVER.fullmatch(value) is None:
        raise SystemExit(f"VERSION is not valid SemVer: {value}")
    return value


def validate_tag(tag: str) -> None:
    expected = f"v{version()}"
    if tag != expected:
        raise SystemExit(f"tag {tag} does not match reviewed VERSION {expected}")
    notes = ROOT / f"RELEASE_NOTES_{tag}.md"
    if not notes.is_file():
        raise SystemExit(f"missing reviewed release notes: {notes.name}")


def spdx_timestamp(epoch: int) -> str:
    return dt.datetime.fromtimestamp(epoch, tz=dt.timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def write_sbom(target: str, output: Path, epoch: int, revision: str) -> None:
    release_version = version()
    document = {
        "spdxVersion": "SPDX-2.3",
        "dataLicense": "CC0-1.0",
        "SPDXID": "SPDXRef-DOCUMENT",
        "name": f"calathea-{release_version}-{target}",
        "documentNamespace": f"https://github.com/hackelia-micrantha/calathea-community/releases/tag/v{release_version}#spdx-{target}-{revision}",
        "creationInfo": {
            "created": spdx_timestamp(epoch),
            "creators": ["Tool: calathea/scripts/release.py"],
        },
        "packages": [
            {
                "name": "calathea",
                "SPDXID": "SPDXRef-Package-Calathea",
                "versionInfo": release_version,
                "downloadLocation": "NOASSERTION",
                "filesAnalyzed": False,
                "licenseConcluded": "NOASSERTION",
                "licenseDeclared": "MPL-2.0",
                "copyrightText": "NOASSERTION",
                "externalRefs": [
                    {
                        "referenceCategory": "PACKAGE-MANAGER",
                        "referenceType": "purl",
                        "referenceLocator": f"pkg:golang/github.com/hackelia-micrantha/calathea-community@{release_version}",
                    }
                ],
            }
        ],
        "relationships": [
            {
                "spdxElementId": "SPDXRef-DOCUMENT",
                "relationshipType": "DESCRIBES",
                "relatedSpdxElement": "SPDXRef-Package-Calathea",
            }
        ],
        "annotations": [
            {
                "annotationDate": spdx_timestamp(epoch),
                "annotationType": "OTHER",
                "annotator": "Tool: calathea/scripts/release.py",
                "comment": "Calathea currently has no third-party runtime Go modules; go.mod contains only the module and Go language version.",
            }
        ],
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps(document, sort_keys=True, separators=(",", ":")) + "\n", encoding="utf-8")


def add_member(tar: tarfile.TarFile, name: str, data: bytes, mode: int, epoch: int) -> None:
    info = tarfile.TarInfo(name=name)
    info.size = len(data)
    info.mode = mode
    info.uid = 0
    info.gid = 0
    info.uname = ""
    info.gname = ""
    info.mtime = epoch
    tar.addfile(info, io.BytesIO(data))


def package(target: str, binary: Path, sbom: Path, dist: Path, epoch: int) -> None:
    release_version = version()
    os_name, arch = target.split("/", 1)
    archive_name = f"calathea_{release_version}_{os_name}_{arch}.tar.gz"
    archive = dist / archive_name
    dist.mkdir(parents=True, exist_ok=True)

    files = [
        ("calathea", binary.read_bytes(), 0o755),
        ("LICENSE", (ROOT / "LICENSE").read_bytes(), 0o644),
        ("README.md", (ROOT / "README.md").read_bytes(), 0o644),
        ("SBOM.spdx.json", sbom.read_bytes(), 0o644),
    ]

    with archive.open("wb") as raw:
        with gzip.GzipFile(filename="", mode="wb", fileobj=raw, mtime=epoch, compresslevel=9) as gz:
            with tarfile.open(fileobj=gz, mode="w", format=tarfile.PAX_FORMAT) as tar:
                for name, data, mode in files:
                    add_member(tar, name, data, mode, epoch)

    digest = hashlib.sha256(archive.read_bytes()).hexdigest()
    checksum = archive.with_suffix(archive.suffix + ".sha256")
    checksum.write_text(f"{digest}  {archive.name}\n", encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser()
    commands = parser.add_subparsers(dest="command", required=True)

    validate = commands.add_parser("validate-tag")
    validate.add_argument("--tag", required=True)

    sbom = commands.add_parser("sbom")
    sbom.add_argument("--target", required=True)
    sbom.add_argument("--output", type=Path, required=True)
    sbom.add_argument("--epoch", type=int, required=True)
    sbom.add_argument("--revision", required=True)

    pkg = commands.add_parser("package")
    pkg.add_argument("--target", required=True)
    pkg.add_argument("--binary", type=Path, required=True)
    pkg.add_argument("--sbom", type=Path, required=True)
    pkg.add_argument("--dist", type=Path, required=True)
    pkg.add_argument("--epoch", type=int, required=True)

    args = parser.parse_args()
    if args.command == "validate-tag":
        validate_tag(args.tag)
    elif args.command == "sbom":
        write_sbom(args.target, args.output, args.epoch, args.revision)
    elif args.command == "package":
        package(args.target, args.binary, args.sbom, args.dist, args.epoch)


if __name__ == "__main__":
    main()
