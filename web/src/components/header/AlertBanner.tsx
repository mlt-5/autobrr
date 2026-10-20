/*
 * Copyright (c) 2021 - 2026, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { useTranslation } from "react-i18next";

import { AlertsQueryOptions } from "@api/queries";

export const AlertBanner = () => {
  const { t } = useTranslation("common");

  const { data: alerts } = useQuery(AlertsQueryOptions());

  const ircNetworks = alerts?.filter((alert) => alert.kind === "IRC_UNHEALTHY") ?? [];
  const lists = alerts?.filter((alert) => alert.kind === "LIST_REFRESH_ERROR") ?? [];

  return (
    <>
      {ircNetworks.length > 0 && (
        <ErrorBanner
          to="/settings/irc"
          label={ircNetworks.length === 1 ? t("header.ircNetworkUnhealthy") : t("header.multipleIrcNetworksUnhealthy")}
          alerts={ircNetworks}
        />
      )}

      {lists.length > 0 && (
        <ErrorBanner
          to="/settings/lists"
          label={lists.length === 1 ? t("header.listRefreshFailed") : t("header.multipleListRefreshesFailed")}
          alerts={lists}
        />
      )}
    </>
  );
};

interface ErrorBannerProps {
  to: "/settings/irc" | "/settings/lists";
  label: string;
  alerts: Alert[];
}

const ErrorBanner = ({ to, label, alerts }: ErrorBannerProps) => (
  <Link to={to} className="flex flex-wrap mt-4 py-2 px-3 gap-y-1 bg-red-500 rounded-sm justify-center items-center cursor-pointer">
    <ExclamationTriangleIcon className="h-6 w-6 text-red-100" />
    <span className="text-red-100 font-medium mx-3">{label}</span>
    <span className="flex flex-wrap justify-center gap-1">
      {alerts.map((alert) => (
        <AlertChip key={`${alert.kind}-${alert.subject_id}`} title={alert.message}>
          {alert.subject}
        </AlertChip>
      ))}
    </span>
  </Link>
);

interface AlertChipProps {
  title?: string;
  children: string;
}

const AlertChip = ({ title, children }: AlertChipProps) => (
  <span
    title={title}
    className="inline-flex items-center rounded-md bg-red-100 px-2.5 py-0.5 text-sm font-medium text-red-800"
  >
    {children}
  </span>
);
