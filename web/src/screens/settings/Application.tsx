/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";

import { ConfigQueryOptions } from "@api/queries";
import { SettingsContext } from "@utils/Context";
import type { Language, Theme } from "@utils/Context";
import { Checkbox } from "@components/Checkbox";

import { Section, RowItem } from "./_components";

function ApplicationSettings() {
  const { t } = useTranslation(["common", "settings"]);
  const [settings, setSettings] = SettingsContext.use();

  const { data } = useQuery(ConfigQueryOptions());

  return (
    <Section
      title={t("settings:application.title")}
      description={t("settings:application.description")}
    >
      <div className="-mx-4 divide-y divide-gray-150 dark:divide-gray-750">
        <form className="mt-6 pb-4" action="#" method="POST">
          {data && (
            <div className="grid grid-cols-12 gap-2 sm:gap-6 px-4 sm:px-6">
              <div className="col-span-12 sm:col-span-6">
                <label htmlFor="host" className="block ml-px text-xs font-bold text-gray-700 dark:text-white uppercase tracking-wide">
                  {t("settings:application.host")}
                </label>
                <input
                  type="text"
                  name="host"
                  id="host"
                  value={data.host}
                  disabled={true}
                  className="mt-1 block w-full sm:text-sm rounded-md border-gray-300 dark:border-gray-750 bg-gray-100 dark:bg-gray-825 dark:text-gray-100"
                />
              </div>

              <div className="col-span-12 sm:col-span-6">
                <label htmlFor="port" className="block ml-px text-xs font-bold text-gray-700 dark:text-white uppercase tracking-wide">
                  {t("settings:application.port")}
                </label>
                <input
                  type="text"
                  name="port"
                  id="port"
                  value={data.port}
                  disabled={true}
                  className="mt-1 block w-full sm:text-sm rounded-md border-gray-300 dark:border-gray-750 bg-gray-100 dark:bg-gray-825 dark:text-gray-100"
                />
              </div>

            </div>
          )}
        </form>

        <RowItem label={t("settings:application.version")} value={data?.version} />
        {data?.commit && <RowItem label={t("settings:application.commit")} value={data.commit} />}
        {data?.date && <RowItem label={t("settings:application.buildDate")} value={data.date} />}
        <RowItem label={t("settings:application.application")} value={data?.application} />
        <RowItem label={t("settings:application.configPath")} value={data?.config_dir} />
        <div className="py-0.5">
          <Checkbox
            name="debug"
            label={t("settings:application.webuiDebugMode")}
            value={settings.debug}
            className="p-4 sm:px-6"
            setValue={
              (newValue: boolean) => setSettings((prevState) => ({
                ...prevState,
                debug: newValue
              }))
            }
          />
        </div>
        <div className="flex items-center justify-between p-4 sm:px-6">
          <div className="flex flex-col mr-4">
            <p className="text-sm font-medium whitespace-nowrap text-gray-900 dark:text-white">
              {t("settings:application.theme")}
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {t("settings:application.themeDescription")}
            </p>
          </div>
          <div>
          <select
            value={settings.theme}
            onChange={(e) => setSettings((prevState) => ({
              ...prevState,
              theme: e.target.value as Theme
            }))}
            className="rounded-md min-w-32 border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 cursor-pointer text-sm text-gray-900 dark:text-gray-100 px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="light">{t("common:theme.light")}</option>
            <option value="dark">{t("common:theme.dark")}</option>
            <option value="system">{t("common:theme.system")}</option>
          </select>
          </div>
        </div>
        <div className="flex items-center justify-between p-4 sm:px-6">
          <div className="flex flex-col mr-4">
            <p className="text-sm font-medium whitespace-nowrap text-gray-900 dark:text-white">
              {t("common:language.label")}
            </p>
            <p className="text-sm text-gray-500 dark:text-gray-400">
              {t("settings:application.languageDescription")}
            </p>
          </div>
          <div>
          <select
            value={settings.language}
            onChange={(e) => setSettings((prevState) => ({
              ...prevState,
              language: e.target.value as Language
            }))}
            className="rounded-md min-w-56 border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 cursor-pointer text-sm text-gray-900 dark:text-gray-100 px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="en">{t("common:language.english")}</option>
            <option value="fr">{t("common:language.french")}</option>
            <option value="de">{t("common:language.german")}</option>
            <option value="cs">{t("common:language.czech")}</option>
            <option value="no">{t("common:language.norwegian")}</option>
            <option value="ru">{t("common:language.russian")}</option>
            <option value="es">{t("common:language.spanish")}</option>
            <option value="zh-CN">{t("common:language.simplifiedChinese")}</option>
          </select>
          </div>
        </div>
      </div>
    </Section>
  );
}

export default ApplicationSettings;
