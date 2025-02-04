import "../index.css";

interface Props {
  children: React.ReactNode;
  title?: string;
  description?: string;
  absoluteNavbar?: boolean;
}

export function Layout({ children }: Props) {
  return (
    <main className="relative flex min-h-screen flex-col overflow-auto">
      {children}
    </main>
  );
}
