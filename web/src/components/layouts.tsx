/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

import { Header } from "@components/header";
import { SettingsContext } from "@utils/Context";

export function AppLayout() {
  return (
    <div className="flex flex-col min-h-screen">
      <Header />
      <Outlet />
    </div>
  )
}

export const RootComponent = () => {
  const debug = SettingsContext.useSelector((s) => s.debug);
  return (
    <div className="flex flex-col min-h-screen">
      <Outlet />
      {debug ? (
        <>
          {import.meta.env.DEV && <TanStackRouterDevtools />}
          {import.meta.env.DEV && <ReactQueryDevtools initialIsOpen={false} />}
        </>
      ) : null}
    </div>
  )
}
