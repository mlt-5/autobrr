/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { WarningAlert } from "@components/alerts";
import { FilterLayout, FilterSection } from "@screens/filters/sections/_components.tsx";
import { TextAreaAutoResize, TextField } from "@components/inputs";
import { useTranslation } from "react-i18next";


export const Test = () => {
  const { t } = useTranslation("filters");

  return (
  <WarningAlert
    alert={t("actionComponents.test.alert")}
    className="mt-2"
    colors="text-fuchsia-700 bg-fuchsia-100 dark:bg-fuchsia-200 dark:text-fuchsia-800"
    text={t("actionComponents.test.text")}
  />
  );
};

export const WatchFolder = ({ idx }: ClientActionProps) => {
  const { t } = useTranslation("filters");

  return (
  <FilterSection
    title={t("actionComponents.watchFolder.title")}
    subtitle={t("actionComponents.watchFolder.subtitle")}
  >
    <FilterLayout>
      <TextAreaAutoResize
        name={`actions[${idx}].watch_folder`}
        label={t("actionComponents.watchFolder.directory")}
        placeholder={t("actionComponents.watchFolder.directoryPlaceholder")}
      />
    </FilterLayout>
  </FilterSection>
  );
};

export const WebHook = ({ idx }: ClientActionProps) => {
  const { t } = useTranslation("filters");

  return (
  <FilterSection
    title={t("actionComponents.webhook.title")}
    subtitle={t("actionComponents.webhook.subtitle")}
  >
    <FilterLayout>
      <TextField
        name={`actions[${idx}].webhook_host`}
        label={t("actionComponents.webhook.endpoint")}
        columns={6}
        placeholder={t("actionComponents.webhook.endpointPlaceholder")}
        tooltip={
          <p>{t("actionComponents.webhook.endpointTooltip")}</p>
        }
      />
    </FilterLayout>
    <TextAreaAutoResize
      name={`actions[${idx}].webhook_data`}
      label={t("actionComponents.webhook.payload")}
      placeholder={t("actionComponents.webhook.payloadPlaceholder")}
    />
  </FilterSection>
  );
};
