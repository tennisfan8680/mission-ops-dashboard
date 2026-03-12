type TopBarProps = {
  role: string;
};

export default function TopBar({ role }: TopBarProps) {
  return (
    <header className="topbar">
      <div>
        <h1>Mission Ops Dashboard</h1>
        <p>Common Operating Picture + Alert Triage</p>
      </div>

      <div className="topbar-right">
        <div className="status-pill online">System Online</div>
        <div className="role-badge">{role}</div>
      </div>
    </header>
  );
}