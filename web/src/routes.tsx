/*
 * Copyright (c) 2021 - 2025, Ludvig Lundgren and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { z } from "zod/mini";
import {
  createRootRouteWithContext,
  createRoute,
  createRouter,
  lazyRouteComponent,
  notFound,
} from "@tanstack/react-router";
import { QueryClient } from "@tanstack/react-query";

import { NotFound } from "@components/alerts/NotFound";
import { FilterNotFound } from "@screens/filters/NotFound";
import { Settings } from "@screens/Settings";
import {
  ConfigQueryOptions,
  FeedsQueryOptions,
  FilterByIdQueryOptions,
  IndexersQueryOptions,
  IrcQueryOptions, ListsQueryOptions,
  NotificationsQueryOptions,
  ProxiesQueryOptions
} from "@api/queries";
import { RingResizeSpinner } from "@components/Icons";
import { queryClient } from "@api/QueryClient";

import { ErrorPage } from "@components/alerts";
import { AppLayout, RootComponent } from "@components/layouts";

const DashboardRoute = createRoute({
  getParentRoute: () => LayoutRoute,
  path: '/',
  loader: () => {
    // https://tanstack.com/router/v1/docs/guide/deferred-data-loading#deferred-data-loading-with-defer-and-await
    // TODO load stats

    // TODO load recent releases

    return {}
  },
  component: lazyRouteComponent(() => import("@screens/Dashboard"), "Dashboard"),
});

const FiltersRoute = createRoute({
  getParentRoute: () => LayoutRoute,
  path: 'filters'
});

const FilterIndexRoute = createRoute({
  getParentRoute: () => FiltersRoute,
  path: '/',
  component: lazyRouteComponent(() => import("@screens/filters/List"), "Filters"),
});

export const FilterGetByIdRoute = createRoute({
  getParentRoute: () => FiltersRoute,
  path: '$filterId',
  params: {
    parse: (params) => {
      if (!/^[1-9]\d*$/.test(params.filterId)) {
        throw notFound();
      }
      const filterId = z.int().safeParse(Number(params.filterId));
      if (!filterId.success) {
        throw notFound();
      }
      return { filterId: filterId.data };
    },
    stringify: ({ filterId }) => ({ filterId: `${filterId}` }),
  },
  loader: async ({ context, params }) => {
    try {
      const filter = await context.queryClient.query({ ...FilterByIdQueryOptions(params.filterId), staleTime: "static" })
      return { filter }
    } catch {
      throw notFound()
    }
  },
  component: lazyRouteComponent(() => import("@screens/filters/Details"), "FilterDetails"),
  notFoundComponent: () => {
    return <FilterNotFound />
  },
});

export const FilterGeneralRoute = createRoute({
  getParentRoute: () => FilterGetByIdRoute,
  path: '/',
  component: lazyRouteComponent(() => import("@screens/filters/sections/General"), "General")
});

export const FilterMoviesTvRoute = createRoute({
  getParentRoute: () => FilterGetByIdRoute,
  path: 'movies-tv',
  component: lazyRouteComponent(() => import("@screens/filters/sections/MoviesAndTV"), "MoviesTv")
});

export const FilterMusicRoute = createRoute({
  getParentRoute: () => FilterGetByIdRoute,
  path: 'music',
  component: lazyRouteComponent(() => import("@screens/filters/sections/Music"), "Music")
});

export const FilterBooksRoute = createRoute({
  getParentRoute: () => FilterGetByIdRoute,
  path: 'books',
  component: lazyRouteComponent(() => import("@screens/filters/sections/Books"), "Books")
});

export const FilterAdvancedRoute = createRoute({
  getParentRoute: () => FilterGetByIdRoute,
  path: 'advanced',
  component: lazyRouteComponent(() => import("@screens/filters/sections/Advanced"), "Advanced")
});

export const FilterExternalRoute = createRoute({
  getParentRoute: () => FilterGetByIdRoute,
  path: 'external',
  component: lazyRouteComponent(() => import("@screens/filters/sections/External"), "External")
});

export const FilterActionsRoute = createRoute({
  getParentRoute: () => FilterGetByIdRoute,
  path: 'actions',
  component: lazyRouteComponent(() => import("@screens/filters/sections/Actions"), "Actions")
});

export const FilterNotificationsRoute = createRoute({
  getParentRoute: () => FilterGetByIdRoute,
  path: 'notifications',
  component: lazyRouteComponent(() => import("@screens/filters/sections/Notifications"), "Notifications")
});

export const ReleasesRoute = createRoute({
  getParentRoute: () => LayoutRoute,
  path: 'releases',
  component: lazyRouteComponent(() => import("@screens/Releases"), "Releases"),
  validateSearch: (search) => z.object({
    page: z.optional(z.number()),
    pageSize: z.optional(z.number()),
    q: z.optional(z.string()),
    action_status: z.optional(z.enum(['PUSH_APPROVED', 'PUSH_REJECTED', 'PENDING', 'PUSH_ERROR', ''])),
    indexer: z.optional(z.string()),
  }).parse(search),
});

export const SettingsRoute = createRoute({
  getParentRoute: () => LayoutRoute,
  path: 'settings',
  pendingMs: 3000,
  component: Settings
});

export const SettingsIndexRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: '/',
  component: lazyRouteComponent(() => import("@screens/settings/Application"))
});

export const SettingsLogRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: 'logs',
  loader: (opts) => opts.context.queryClient.query({ ...ConfigQueryOptions(), staleTime: "static" }),
  component: lazyRouteComponent(() => import("@screens/settings/Logs"))
});

export const SettingsIndexersRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: 'indexers',
  loader: (opts) => opts.context.queryClient.query({ ...IndexersQueryOptions(), staleTime: "static" }),
  component: lazyRouteComponent(() => import("@screens/settings/Indexer"))
});

export const SettingsIrcRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: 'irc',
  loader: (opts) => opts.context.queryClient.query({ ...IrcQueryOptions(), staleTime: "static" }),
  component: lazyRouteComponent(() => import("@screens/settings/Irc"))
});

export const SettingsListsRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: 'lists',
  loader: (opts) => opts.context.queryClient.query({ ...ListsQueryOptions(), staleTime: "static" }),
  component: lazyRouteComponent(() => import("@screens/settings/Lists"))
});

export const SettingsFeedsRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: 'feeds',
  loader: (opts) => opts.context.queryClient.query({ ...FeedsQueryOptions(), staleTime: "static" }),
  component: lazyRouteComponent(() => import("@screens/settings/Feed"))
});

export const SettingsNotificationsRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: 'notifications',
  loader: (opts) => opts.context.queryClient.query({ ...NotificationsQueryOptions(), staleTime: "static" }),
  component: lazyRouteComponent(() => import("@screens/settings/Notifications"))
});

export const SettingsProxiesRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: 'proxies',
  loader: (opts) => opts.context.queryClient.query({ ...ProxiesQueryOptions(), staleTime: "static" }),
  component: lazyRouteComponent(() => import("@screens/settings/Proxy"))
});

export const SettingsReleasesRoute = createRoute({
  getParentRoute: () => SettingsRoute,
  path: 'releases',
  component: lazyRouteComponent(() => import("@screens/settings/Releases"))
});

export const LogsRoute = createRoute({
  getParentRoute: () => LayoutRoute,
  path: 'logs',
  component: lazyRouteComponent(() => import("@screens/Logs"), "Logs")
});

export const NotificationInboxRoute = createRoute({
  getParentRoute: () => LayoutRoute,
  path: 'notifications',
  component: lazyRouteComponent(() => import("@screens/NotificationInbox"), "NotificationInbox"),
  validateSearch: (search) => z.object({
    page: z.catch(z.optional(z.int().check(z.nonnegative())), undefined),
    pageSize: z.catch(z.optional(z.literal([10, 25, 50, 100])), undefined),
    unread: z.optional(z.boolean()),
    event: z.optional(z.enum(['PUSH_ERROR', 'PUSH_REJECTED', 'PUSH_APPROVED', 'IRC_DISCONNECTED', 'IRC_RECONNECTED', 'IRC_UNHEALTHY', 'IRC_HEALTHY', 'LIST_REFRESH_ERROR', 'LIST_REFRESH_SUCCESS', 'FEED_REFRESH_ERROR', 'FEED_REFRESH_SUCCESS'])),
  }).parse(search),
});

export const LayoutRoute = createRoute({
  getParentRoute: () => RootRoute,
  component: AppLayout,
  id: 'layout',
});

export const RootRoute = createRootRouteWithContext<{
  queryClient: QueryClient
}>()({
  component: RootComponent,
  notFoundComponent: NotFound,
});

const filterRouteTree = FiltersRoute.addChildren([FilterIndexRoute, FilterGetByIdRoute.addChildren([FilterGeneralRoute, FilterMoviesTvRoute, FilterMusicRoute, FilterBooksRoute, FilterAdvancedRoute, FilterExternalRoute, FilterActionsRoute, FilterNotificationsRoute])])
const settingsRouteTree = SettingsRoute.addChildren([SettingsIndexRoute, SettingsLogRoute, SettingsIndexersRoute, SettingsIrcRoute, SettingsListsRoute, SettingsFeedsRoute, SettingsNotificationsRoute, SettingsProxiesRoute, SettingsReleasesRoute])
const routeTree = RootRoute.addChildren([
  LayoutRoute.addChildren([DashboardRoute, filterRouteTree, ReleasesRoute, settingsRouteTree, LogsRoute, NotificationInboxRoute])
]);

export const Router = createRouter({
  routeTree,
  defaultPreload: "intent",
  defaultPendingComponent: () => (
    <div className="flex grow items-center justify-center col-span-9">
      <RingResizeSpinner className="text-blue-500 size-24" />
    </div>
  ),
  defaultErrorComponent: (ctx) => (
    <ErrorPage error={ctx.error} reset={ctx.reset} />
  ),
  context: {
    queryClient
  },
});

