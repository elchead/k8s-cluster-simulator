from regex import D
from job import *


def test_get_migration_time():
    assert 157 == get_migration_time(50.0)


def test_poddata_migration():
    sut = PodData.withdata([10, 20, 30], [50, 40, 50], [], is_migrated=True)
    assert 10 == sut.get_migration_timestamp()
    assert 50 == sut.get_migration_size()


# def test_cnstruct_migration_plot_data():
#     t = 100
#     duration = 500
#     sz = 50
#     tick_interval = 100
#     x, y = create_migration_blocker_data(t, duration, sz, tick_interval)
#     assert ([100, 200, 300, 400, 500, 600] == x).all()
#     assert ([50, 50, 50, 50, 50, 50] == y).all()


# DEPRECATED: using time from sim log (more accurate)
# class TestMigrationTime:
#     job = Job()
#     job.add_pod("owo", "zone2", PodData.withdata([10, 20, 30], [1, 2, 50], []))
#     job.add_pod("mowo", "zone2", PodData.withdata([10, 20, 30], [50, 5, 6], []))

#     def test_get_single_migration_time(self):
#         assert self.job.get_migration_duration() == 168.0

#     def test_add_pod_migrations(self):
#         self.job.add_pod("mmowo", "zone3", PodData.withdata([10, 20, 30], [100, 5, 6], []))
#         assert self.job.get_migration_duration() == 168.0 * 3


def test_get_job_execution():
    job = Job()
    job.add_pod("owo", "zone2", PodData.withdata([10, 20, 30], [1, 2, 3], []))
    job.add_pod("mowo", "zone2", PodData.withdata([10, 20, 30], [4, 5, 6], []))
    assert job.get_execution_time() == 60


class TestGetPodRuns:
    job = Job()
    job.add_pod("mowo", "zone3", PodData.withdata([10, 20, 30], [4, 5, 6], []))
    job.add_pod("owo", "zone2", PodData.withdata([10, 20, 30], [1, 2, 3], []))
    job2 = Job()
    job2.add_pod("oza", "zone2", PodData.withdata([10, 20, 30], [1, 2, 3], []))
    it = tuple(job.get_pod_runs_for_plot())
    it2 = tuple(job2.get_pod_runs_for_plot())

    def test_first_item(self):
        zone, poddata = self.it[0]
        assert "zone2" == zone
        assert [1, 2, 3] == poddata.memory
        assert [10, 20, 30] == poddata.time

    def test_second_item(self):
        zone, poddata = self.it[1]
        assert "zone3" == zone
        assert [4, 5, 6] == poddata.memory
        assert np.array_equal([40, 50, 60], poddata.time) == True

    def test_time_is_not_mutated(self):
        assert self.job.node_data[1].time == [10, 20, 30]

    def test_nbr_migrations(self):
        assert self.job.nbr_migrations == 1

    def test_migrated_pod_is_marked_at_end(self):
        zone, poddata = self.it[0]
        assert [2] == poddata.migration_idx

    def test_restarted_pod_is_marked_at_beginning(self):
        zone, poddata = self.it[1]
        assert [0] == poddata.migration_idx

    def test_not_migrated_pod_is_not_marked(self):
        zone, poddata = self.it2[0]
        assert [] == poddata.migration_idx

