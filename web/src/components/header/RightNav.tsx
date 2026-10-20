/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { useTranslation } from "react-i18next";

import { InboxMenu } from "./InboxMenu";

import { MoonIcon, SunIcon, ComputerDesktopIcon } from "@heroicons/react/24/outline";
import { SettingsContext, isDarkTheme } from "@utils/Context";
import type { Theme } from "@utils/Context";

export const RightNav = () => {
  const { t } = useTranslation("common");
  const [settings, setSettings] = SettingsContext.use();

  const nextTheme: Record<Theme, Theme> = {
    light: "dark",
    dark: "system",
    system: "light"
  };

  const themeIcons: Record<Theme, typeof MoonIcon> = {
    dark: MoonIcon,
    light: SunIcon,
    system: ComputerDesktopIcon
  };

  const themeLabels: Record<Theme, string> = {
    light: t("theme.toggleLightToDark"),
    dark: t("theme.toggleDarkToSystem"),
    system: t("theme.toggleSystemToLight")
  };

  const toggleTheme = () => {
    setSettings(prevState => ({
      ...prevState,
      theme: nextTheme[prevState.theme]
    }));
  };

  const ThemeIcon = themeIcons[settings.theme];

  return (
    <div className="hidden sm:block">
      <div className="ml-4 flex items-center sm:ml-6">
        <div className="mr-2">
          <InboxMenu />
        </div>
        <div className="mt-1 items-center">
          <button
            onClick={toggleTheme}
            className="p-1 rounded-full hover:cursor-pointer focus:outline-hidden focus:none transition duration-100 ease-out transform hover:bg-gray-200 dark:hover:bg-gray-800 hover:scale-100"
            title={themeLabels[settings.theme]}
          >
            <ThemeIcon className={`h-4 w-4 transition duration-100 ease-out transform ${isDarkTheme(settings.theme) ? "text-gray-500" : "text-gray-600"}`} aria-hidden="true" />
          </button>
        </div>
      </div>
    </div>
  );
};
