import { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { cn } from "@/lib/utils";
import { Toaster } from "@/components/ui/sonner";

const topNavItems = [
    { to: '/php', label: 'PHP' },
    { to: '/sites', label: 'Sites' },
    { to: '/services', label: 'Local Services' },
    { to: '/dumps', label: 'Live Dumps' },
];

const bottomNavItems = [
    { to: '/settings', label: 'Settings' },
];

export function BaseLayout({ children }: { children: ReactNode }) {
    return (
        <div className="h-screen w-full overflow-hidden text-foreground selection:bg-primary selection:text-primary-foreground">
            <div className="flex h-full w-full min-h-0 bg-card">
                <aside className="flex w-full shrink-0 flex-col border-b-2 border-border bg-sidebar p-4 md:w-64 md:border-r-2 md:border-b-0 md:px-5 md:py-6 dark:border-zinc-800 dark:bg-zinc-950/70">
                    <nav className="overflow-x-auto md:overflow-visible">
                        <ul className="flex min-w-max gap-2 md:min-w-0 md:flex-col md:gap-1">
                            {topNavItems.map((item) => (
                                <li key={item.to} className="shrink-0 md:shrink">
                                    <NavLink
                                        to={item.to}
                                        className={({ isActive }) =>
                                            cn(
                                                "block border-l-2 border-transparent px-3 py-2 text-left text-xs font-bold tracking-[0.12em] uppercase transition-all font-mono",
                                                isActive
                                                    ? "border-primary bg-primary/10 text-primary"
                                                    : "text-muted-foreground hover:bg-muted hover:text-foreground",
                                            )
                                        }
                                    >
                                        <span className="pr-1 text-primary/90">&gt;_</span>
                                        {item.label}
                                    </NavLink>
                                </li>
                            ))}
                        </ul>
                    </nav>

                    <div className="mt-4 border-t-2 border-border pt-3 md:mt-auto dark:border-zinc-800">
                        <p className="mb-2 font-mono text-[10px] tracking-[0.12em] text-muted-foreground uppercase">Settings + Diagnostics</p>
                        <nav>
                            <ul className="flex min-w-max gap-2 md:min-w-0 md:flex-col md:gap-1">
                                {bottomNavItems.map((item) => (
                                    <li key={item.to} className="shrink-0 md:shrink">
                                        <NavLink
                                            to={item.to}
                                            className={({ isActive }) =>
                                                cn(
                                                    "block border-l-2 border-transparent px-3 py-2 text-left text-xs font-bold tracking-[0.12em] uppercase transition-all font-mono",
                                                    isActive
                                                        ? "border-primary bg-primary/10 text-primary"
                                                        : "text-muted-foreground hover:bg-muted hover:text-foreground",
                                                )
                                            }
                                        >
                                            <span className="pr-1 text-primary/90">&gt;_</span>
                                            {item.label}
                                        </NavLink>
                                    </li>
                                ))}
                            </ul>
                        </nav>
                    </div>
                </aside>

                <main className="min-h-0 flex-1 overflow-y-auto bg-card p-5 md:p-8">
                    <div className="relative mx-auto max-w-6xl">
                        {children}
                    </div>
                </main>
            </div>
            <Toaster />
        </div>
    );
}
