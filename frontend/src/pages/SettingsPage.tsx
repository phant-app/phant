import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useTheme } from "@/components/theme-provider";

export function SettingsPage() {
    const { theme, setTheme } = useTheme();

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
                <p className="text-muted-foreground mt-2">
                    Manage your app preferences and appearance.
                </p>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Appearance</CardTitle>
                    <CardDescription>Customize the look and feel of Phant.</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex gap-4">
                        <Button 
                            variant={theme === "light" ? "default" : "outline"} 
                            onClick={() => setTheme("light")}
                        >
                            Light
                        </Button>
                        <Button 
                            variant={theme === "dark" ? "default" : "outline"} 
                            onClick={() => setTheme("dark")}
                        >
                            Dark
                        </Button>
                        <Button 
                            variant={theme === "system" ? "default" : "outline"} 
                            onClick={() => setTheme("system")}
                        >
                            System
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
