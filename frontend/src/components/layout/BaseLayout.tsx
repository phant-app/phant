import { ReactNode } from "react";
import { NavLink } from "react-router-dom";
import { cn } from "@/lib/utils";
import logoUrl from "@/assets/logo.png";

const navGroups = [
    {
        title: 'Product',
        items: [
            { to: '/php', label: 'PHP Manager' },
            { to: '/valet', label: 'Valet Linux' },
            { to: '/sites', label: 'Sites' },
            { to: '/services', label: 'Local Services' },
        ],
    },
    {
        title: 'Tools',
        items: [
            { to: '/dumps', label: 'Live Dumps' },
            { to: '/settings', label: 'Settings' },
        ],
    },
];

export function BaseLayout({ children }: { children: ReactNode }) {
    return (
        <div className="flex h-screen w-full bg-background text-foreground selection:bg-primary/30 selection:text-primary min-h-screen">
            <aside className="w-64 shrink-0 border-r border-border bg-card flex flex-col">
                <div className="p-6 flex items-center">
                    <img src={logoUrl} alt="Phant Logo" className="transition hover:animate-tilt duration-600" />
                </div>
                
                <nav className="flex-1 px-4 space-y-6 overflow-y-auto pt-2 pb-6">
                    {navGroups.map((group) => (
                        <div key={group.title}>
                            <h3 className="mb-2 px-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                                {group.title}
                            </h3>
                            <ul className="space-y-1">
                                {group.items.map((item) => (
                                    <li key={item.to}>
                                        <NavLink
                                            to={item.to}
                                            className={({ isActive }) =>
                                                cn(
                                                    "flex items-center rounded-md px-2 py-1.5 text-sm font-medium transition-colors",
                                                    isActive
                                                        ? "bg-primary/10 text-primary"
                                                        : "text-muted-foreground hover:bg-muted hover:text-foreground"
                                                )
                                            }
                                        >
                                            {item.label}
                                        </NavLink>
                                    </li>
                                ))}
                            </ul>
                        </div>
                    ))}
                </nav>
            </aside>

            <main className="flex-1 overflow-y-auto">
                <div className="p-8 max-w-5xl mx-auto space-y-6">
                    {children}
                </div>
            </main>
        </div>
    );
}
