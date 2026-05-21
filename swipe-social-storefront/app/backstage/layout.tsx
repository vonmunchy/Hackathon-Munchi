export default function BackstageLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-dvh bg-slate-900 text-white">
      {children}
    </div>
  );
}
