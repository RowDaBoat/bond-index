import os
import shutil
import signal
import subprocess
import tempfile
import time
import json

from contextlib import contextmanager
from pathlib import Path
from typing import Callable, Iterable, List, Optional


SCRIPT_DIR = Path(__file__).resolve().parent
PROJECT_DIR = SCRIPT_DIR.parent

ORD_PORT = int(os.environ.get("ORD_PORT", "4080"))
BOND_PORT = int(os.environ.get("BOND_PORT", "8080"))


class TestEnvironment:
    def __init__(self, debug: bool = False, debug_bitcoind: bool = False, debug_ord: bool = False, debug_bond: bool = False):
        self._tempdir = tempfile.mkdtemp()
        self.test_dir = Path(self._tempdir)
        self.bitcoin_datadir = self.test_dir / "bitcoin"
        self.ord_datadir = self.test_dir / "ord"
        self.bond_datadir = self.test_dir / "bond-data"

        self.bitcoin_datadir.mkdir(parents=True, exist_ok=True)
        self.ord_datadir.mkdir(parents=True, exist_ok=True)
        self.bond_datadir.mkdir(parents=True, exist_ok=True)

        if debug:
            debug_bitcoind = debug_ord = debug_bond = True

        self._log_bitcoind = None if debug_bitcoind else subprocess.DEVNULL
        self._log_ord = None if debug_ord else subprocess.DEVNULL
        self._log_bond = None if debug_bond else subprocess.DEVNULL

        self._bitcoind_proc: Optional[subprocess.Popen] = None
        self._ord_proc: Optional[subprocess.Popen] = None
        self._bond_proc: Optional[subprocess.Popen] = None

    # ----------------------------
    # Process helpers
    # ----------------------------

    def _popen(self, args: List[str], stdout=None, stderr=None) -> subprocess.Popen:
        return subprocess.Popen(args, stdout=stdout, stderr=stderr)

    def _run(self, args: List[str], **kwargs) -> subprocess.CompletedProcess:
        return subprocess.run(args, check=True, text=True, **kwargs)

    # ----------------------------
    # CLI wrappers
    # ----------------------------

    def bitcoin_cli(self, *args: str, capture_output: bool = False) -> str:
        cmd = ["bitcoin-cli", "-regtest", f"-datadir={self.bitcoin_datadir}", *args]
        if capture_output:
            result = self._run(cmd, capture_output=True)
            return result.stdout

        self._run(cmd, stdout=self._log_bitcoind, stderr=self._log_bitcoind)
        return ""

    def ord_wallet(self, *args: str, capture_output: bool = False) -> str:
        cmd = [
            "ord",
            "--regtest",
            f"--bitcoin-data-dir={self.bitcoin_datadir}",
            f"--data-dir={self.ord_datadir}",
            "wallet",
            "--server-url",
            f"http://localhost:{ORD_PORT}",
            *args,
        ]
        if capture_output:
            result = self._run(cmd, capture_output=True)
            return result.stdout

        self._run(cmd, stdout=self._log_ord, stderr=self._log_ord)
        return ""

    def ord_inscribe(
        self,
        file_path,
        *,
        destination: Optional[str] = None,
        parent: Optional[str] = None,
        capture_output: bool = False,
    ) -> str:
        base_args = [
            "inscribe",
            "--fee-rate",
            "1",
            "--no-backup",
            "--file",
            str(file_path),
        ]
        if destination is not None:
            base_args.extend(["--destination", destination])
        if parent is not None:
            base_args.extend(["--parent", parent])
        return self.ord_wallet(*base_args, capture_output=capture_output)

    def bond_client(self, path: str, capture_output: bool = True) -> str:
        url = f"http://localhost:{BOND_PORT}{path}"
        cmd = ["curl", "-s", url]
        result = self._run(cmd, capture_output=True)
        return result.stdout if capture_output else ""

    def bond_sync(self) -> None:
        bond_bin = self.test_dir / "bond"
        cmd = [str(bond_bin), "sync", "--data-dir", str(self.bond_datadir / "data")]
        self._run(cmd, stdout=self._log_bond, stderr=self._log_bond)

    # ----------------------------
    # Mining & sync
    # ----------------------------

    def mine(self, blocks: int = 1, address: Optional[str] = None) -> None:
        if address is None:
            # Generate a fresh address from ord wallet
            out = self.ord_wallet("receive", capture_output=True)
            # Very small inline JSON parsing to avoid extra deps
            import json

            data = json.loads(out)
            address = data["addresses"][0]

        self.bitcoin_cli("generatetoaddress", str(blocks), address, capture_output=False)

    def ord_sync(self, timeout: int = 30) -> None:
        # Wait until ord's reported blockcount >= bitcoind's blockcount
        import json
        import urllib.request

        height = int(json.loads(self.bitcoin_cli("getblockcount", capture_output=True)))
        url = f"http://localhost:{ORD_PORT}/blockcount"

        deadline = time.time() + timeout
        while time.time() < deadline:
            try:
                with urllib.request.urlopen(url, timeout=2) as resp:
                    content = resp.read().decode("utf-8")
                    ord_height = int(content)
                    if ord_height >= height:
                        return
            except Exception:
                pass
            time.sleep(1)
        raise RuntimeError(f"ord failed to sync to block {height} within {timeout} seconds")

    def mine_and_sync(self, blocks: int = 1) -> None:
        self.mine(blocks)
        self.ord_sync()
        self.bond_sync()

    # ----------------------------
    # Wallet setup
    # ----------------------------

    def _setup_wallet(self) -> None:
        # Create and fund wallet once per environment
        self.ord_wallet("create")
        receive_out = self.ord_wallet("receive", capture_output=True)

        receive_data = json.loads(receive_out)
        wallet_address = receive_data["addresses"][0]

        self.mine(101, wallet_address)
        self.ord_sync()

    # ----------------------------
    # Assertions
    # ----------------------------

    @staticmethod
    def assert_equals(description: str, expected: str, actual: str) -> None:
        if actual == expected:
            print(f"✅ PASS: {description}")
        else:
            print(f"❌ FAIL: {description}")
            print(f"  expected: {expected}")
            print(f"  actual:   {actual}")
            raise AssertionError(description)

    @staticmethod
    def assert_success(description: str, func: Callable, *args, **kwargs) -> None:
        try:
            func(*args, **kwargs)
            print(f"✅ PASS: {description}")
        except Exception as e:
            print(f"❌ FAIL: {description}")
            raise AssertionError(f"{description}: {e}") from e

    # ----------------------------
    # Lifecycle
    # ----------------------------

    def _build_bond_binary(self) -> Path:
        bond_bin = self.test_dir / "bond"
        build_cmd = ["bash", "-c", f'(cd "{PROJECT_DIR}" && go build -o "{bond_bin}" .)']
        print("Building bond... ", end="", flush=True)
        try:
            self._run(build_cmd, stdout=self._log_bond, stderr=self._log_bond)
        except Exception as e:
            print("failed.")
            raise RuntimeError(f"Failed to build bond: {e}") from e
        else:
            print("ok.")
        return bond_bin

    def _start_service(self, name: str, cmd: list[str], check: Callable[[], bool], log_target) -> subprocess.Popen:
        print(f"Starting {name}... ", end="", flush=True)
        proc = self._popen(cmd, stdout=log_target, stderr=log_target)
        try:
            self._wait_for_start(name, check)
        except Exception:
            print("failed.")
            raise
        else:
            print("ok.")
        return proc

    def start_services(self) -> None:
        # Check dependencies
        for cmd in ("bitcoind", "bitcoin-cli", "ord"):
            if shutil.which(cmd) is None:
                raise RuntimeError(f"{cmd} is not installed or not on PATH")

        # Build bond binary
        bond_bin = self._build_bond_binary()

        # Start bitcoind
        bitcoind_cmd = [
            "bitcoind",
            "-regtest",
            "-txindex",
            f"-datadir={self.bitcoin_datadir}",
        ]
        self._bitcoind_proc = self._start_service("bitcoind", bitcoind_cmd, self._check_bitcoind, self._log_bitcoind)

        # Start ord
        ord_cmd = [
            "ord",
            "--regtest",
            f"--bitcoin-data-dir={self.bitcoin_datadir}",
            f"--data-dir={self.ord_datadir}",
            "server",
            "--http-port",
            str(ORD_PORT),
        ]
        self._ord_proc = self._start_service("ord", ord_cmd, self._check_ord, self._log_ord)

        # Start bond
        bond_cmd = [
            str(bond_bin),
            "--rest-listen-url",
            f"0.0.0.0:{BOND_PORT}",
            "--ord-url",
            f"http://localhost:{ORD_PORT}",
            "--data-dir",
            str(self.bond_datadir / "data"),
            "--start-block",
            "0",
            "--non-interactive",
            "--no-auto-index",
        ]
        self._bond_proc = self._start_service("bond", bond_cmd, self._check_bond, self._log_bond)

        print("All services running.")
        print("")

    def _check_bitcoind(self) -> bool:
        try:
            self.bitcoin_cli("getblockchaininfo", capture_output=False)
            return True
        except Exception:
            return False

    def _check_ord(self) -> bool:
        import urllib.request

        try:
            with urllib.request.urlopen(f"http://localhost:{ORD_PORT}/status", timeout=1):
                return True
        except Exception:
            return False

    def _check_bond(self) -> bool:
        import urllib.request

        try:
            with urllib.request.urlopen(f"http://localhost:{BOND_PORT}/health", timeout=1):
                return True
        except Exception:
            return False

    def _wait_for_start(self, name: str, check: Callable[[], bool], max_attempts: int = 30) -> None:
        for _ in range(max_attempts):
            if check():
                return
            time.sleep(1)
        raise RuntimeError(f"{name} failed to start after {max_attempts} seconds.")

    def cleanup(self) -> None:
        print()
        # Stop bond
        print("Stopping bond... ", end="", flush=True)
        if self._bond_proc is not None:
            self._bond_proc.send_signal(signal.SIGINT)
            try:
                self._bond_proc.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self._bond_proc.kill()
        print("ok.")

        # Stop ord
        print("Stopping ord... ", end="", flush=True)
        if self._ord_proc is not None:
            self._ord_proc.send_signal(signal.SIGINT)
            try:
                self._ord_proc.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self._ord_proc.kill()
        print("ok.")

        # Stop bitcoind
        print("Stopping bitcoind... ", end="", flush=True)
        try:
            self.bitcoin_cli("stop", capture_output=False)
        except Exception:
            pass
        if self._bitcoind_proc is not None:
            try:
                self._bitcoind_proc.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self._bitcoind_proc.kill()
        print("ok.")

        shutil.rmtree(self.test_dir, ignore_errors=True)
        print("All services stopped.")


@contextmanager
def test_environment(*args, **kwargs):
    """
    Context manager that mirrors the test_setup.sh lifecycle:

        with test_environment() as env:
            env.start_services()
            ...
    """
    env = TestEnvironment(*args, **kwargs)
    try:
        env.start_services()
        env._setup_wallet()
        yield env
    finally:
        env.cleanup()

